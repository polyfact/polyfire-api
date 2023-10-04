package transcription

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/google/uuid"
	router "github.com/julienschmidt/httprouter"
	supa "github.com/nedpals/supabase-go"

	db "github.com/polyfact/api/db"
	stt "github.com/polyfact/api/stt"
	"github.com/polyfact/api/utils"
)

func SplitFile(file io.Reader) ([]io.Reader, int, func(), error) {
	id := "split_transcribe-" + uuid.New().String()
	_ = os.Mkdir("/tmp/"+id, 0700)
	close_func := func() {
		os.RemoveAll("/tmp/" + id)
	}
	fmt.Println(id)

	f, err := os.Create("/tmp/" + id + "/audio-file")
	if err != nil {
		return nil, 0, nil, err
	}

	_, err = io.Copy(f, file)
	if err != nil {
		return nil, 0, nil, err
	}

	err = exec.Command("ffmpeg", "-i", "/tmp/"+id+"/audio-file", "/tmp/"+id+"/audio-file.ts", "-ar", "44100").Run()
	if err != nil {
		return nil, 0, nil, err
	}

	ffprobe_result, err := exec.Command("ffprobe", "-i", "/tmp/"+id+"/audio-file", "-show_entries", "format=duration", "-v", "quiet", "-of", "csv=p=0").
		Output()
	if err != nil {
		return nil, 0, nil, err
	}

	duration_ffprobe := strings.Split(strings.Trim(string(ffprobe_result), " \t\n"), ".")

	duration_minutes, err := strconv.Atoi(duration_ffprobe[0])
	if err != nil {
		return nil, 0, nil, err
	}

	duration_seconds, err := strconv.Atoi(duration_ffprobe[1][0:2])
	if err != nil {
		return nil, 0, nil, err
	}

	duration_seconds = duration_minutes*60 + duration_seconds + 1
	os.Remove("/tmp/" + id + "/audio-file")

	split_cmd := exec.Command("split", "-b", "20971400", "/tmp/"+id+"/audio-file.ts")
	split_cmd.Dir = "/tmp/" + id
	err = split_cmd.Run()
	if err != nil {
		return nil, 0, nil, err
	}

	os.Remove("/tmp/" + id + "/audio-file.ts")
	files, err := os.ReadDir("/tmp/" + id)
	if err != nil {
		fmt.Println("readdir")
		return nil, 0, nil, err
	}

	var res []io.Reader = make([]io.Reader, 0)
	for _, file := range files {
		if file.Name() != "audio-file" && file.Name() != "audio-file.mpeg" {
			err := exec.Command("ffmpeg", "-i", "/tmp/"+id+"/"+file.Name(), "/tmp/"+id+"/"+file.Name()+".mp3").Run()
			if err != nil {
				return nil, 0, nil, err
			}

			os.Remove("/tmp/" + id + "/" + file.Name())
			audio_part_r, err := os.Open("/tmp/" + id + "/" + file.Name() + ".mp3")
			if err != nil {
				fmt.Println("open part")
				return nil, 0, nil, err
			}
			res = append(res, audio_part_r)
			fmt.Println("adding:", file.Name())
		}
	}

	return res, duration_seconds, close_func, nil
}

func DownloadFromBucket(bucket string, path string) ([]byte, error) {
	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	supabase := supa.CreateClient(supabaseUrl, supabaseKey)

	return supabase.Storage.From(bucket).Download(path)
}

type TranscribeRequestBody struct {
	FilePath string `json:"file_path"`
}

type Result struct {
	Text string `json:"text"`
}

func Transcribe(w http.ResponseWriter, r *http.Request, _ router.Params) {
	ctx := r.Context()
	user_id := ctx.Value(utils.ContextKeyUserID).(string)
	record := ctx.Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)

	rateLimitStatus := ctx.Value(utils.ContextKeyRateLimitStatus)

	if rateLimitStatus == db.RateLimitStatusReached {
		utils.RespondError(w, record, "rate_limit_reached")
		return
	}

	if rateLimitStatus == db.RateLimitStatusProjectReached {
		utils.RespondError(w, record, "project_rate_limit_reached")
		return
	}

	content_type := r.Header.Get("Content-Type")
	var file_buf_reader io.Reader

	if content_type == "application/json" {
		var input TranscribeRequestBody

		err := json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			utils.RespondError(w, record, "invalid_json")
			return
		}

		b, err := DownloadFromBucket("audio_transcribes", input.FilePath)
		if err != nil {
			fmt.Println(err)
			utils.RespondError(w, record, "read_error")
			return
		}

		file_buf_reader = bytes.NewReader(b)
	} else {
		_, p, _ := mime.ParseMediaType(content_type)
		boundary := p["boundary"]
		reader := multipart.NewReader(r.Body, boundary)
		part, err := reader.NextPart()
		if err == io.EOF {
			utils.RespondError(w, record, "missing_content")
			return
		}
		if err != nil {
			utils.RespondError(w, record, "read_error", err.Error())
			return
		}
		file_buf_reader = bufio.NewReader(part)
	}

	total_str := ""
	// The format doesn't seem to really matter
	files, duration, close_func, err := SplitFile(file_buf_reader)
	if err != nil {
		fmt.Println(err)
		utils.RespondError(w, record, "splitting_error")
	}
	defer close_func()
	for i, r := range files {
		res, err := stt.Transcribe(ctx, r, "mpeg")
		if err != nil {
			fmt.Printf("%v\n", err)
			utils.RespondError(w, record, "transcription_error")
			return
		}
		total_str += " " + res.Text
		fmt.Printf("Transcription %v/%v\n", i+1, len(files))
	}

	db.LogRequestsCredits(user_id, "openai", "whisper", duration*1000, 0, 0, "transcription")

	res := Result{Text: total_str}

	response, _ := json.Marshal(&res)
	record(string(response))

	_ = json.NewEncoder(w).Encode(res)
}
