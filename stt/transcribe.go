package stt

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

	db "github.com/polyfire/api/db"
	providers "github.com/polyfire/api/stt/providers"
	"github.com/polyfire/api/utils"
)

func SplitFile(file io.Reader) ([]io.Reader, int, func(), error) {
	id := "split_transcribe-" + uuid.New().String()
	_ = os.Mkdir("/tmp/"+id, 0700)
	closeFunc := func() {
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

	ffprobeResult, err := exec.Command("ffprobe", "-i", "/tmp/"+id+"/audio-file", "-show_entries", "format=duration", "-v", "quiet", "-of", "csv=p=0").
		Output()
	if err != nil {
		return nil, 0, nil, err
	}

	durationFfprobe := strings.Split(strings.Trim(string(ffprobeResult), " \t\n"), ".")

	durationMinutes, err := strconv.Atoi(durationFfprobe[0])
	if err != nil {
		return nil, 0, nil, err
	}

	durationSeconds, err := strconv.Atoi(durationFfprobe[1][0:2])
	if err != nil {
		return nil, 0, nil, err
	}

	durationSeconds = durationMinutes*60 + durationSeconds + 1
	os.Remove("/tmp/" + id + "/audio-file")

	splitCmd := exec.Command("split", "-b", "20971400", "/tmp/"+id+"/audio-file.ts")
	splitCmd.Dir = "/tmp/" + id
	err = splitCmd.Run()
	if err != nil {
		return nil, 0, nil, err
	}

	os.Remove("/tmp/" + id + "/audio-file.ts")
	files, err := os.ReadDir("/tmp/" + id)
	if err != nil {
		fmt.Println("readdir")
		return nil, 0, nil, err
	}

	var res = make([]io.Reader, 0)
	for _, file := range files {
		if file.Name() != "audio-file" && file.Name() != "audio-file.mpeg" {
			err := exec.Command("ffmpeg", "-i", "/tmp/"+id+"/"+file.Name(), "/tmp/"+id+"/"+file.Name()+".mp3").Run()
			if err != nil {
				return nil, 0, nil, err
			}

			os.Remove("/tmp/" + id + "/" + file.Name())
			audioPartR, err := os.Open("/tmp/" + id + "/" + file.Name() + ".mp3")
			if err != nil {
				fmt.Println("open part")
				return nil, 0, nil, err
			}
			res = append(res, audioPartR)
			fmt.Println("adding:", file.Name())
		}
	}

	return res, durationSeconds, closeFunc, nil
}

func DownloadFromBucket(bucket string, path string) ([]byte, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	supabase := supa.CreateClient(supabaseURL, supabaseKey)

	return supabase.Storage.From(bucket).Download(path)
}

type TranscribeRequestBody struct {
	FilePath string `json:"file_path"`
	Provider string `json:"provider"`
}

type Result struct {
	Text string `json:"text"`
}

func Transcribe(w http.ResponseWriter, r *http.Request, _ router.Params) {
	ctx := r.Context()
	userID := ctx.Value(utils.ContextKeyUserID).(string)
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

	contentType := r.Header.Get("Content-Type")
	var fileBufReader io.Reader

	var providerName = ""
	if contentType == "application/json" {
		var input TranscribeRequestBody

		err := json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			utils.RespondError(w, record, "invalid_json")
			return
		}

		providerName = input.Provider
		b, err := DownloadFromBucket("audio_transcribes", input.FilePath)
		if err != nil {
			fmt.Println(err)
			utils.RespondError(w, record, "read_error")
			return
		}

		fileBufReader = bytes.NewReader(b)
	} else {
		_, p, _ := mime.ParseMediaType(contentType)
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
		fileBufReader = bufio.NewReader(part)
	}

	totalStr := ""
	// The format doesn't seem to really matter
	files, duration, closeFunc, err := SplitFile(fileBufReader)
	if err != nil {
		fmt.Println(err)
		utils.RespondError(w, record, "splitting_error")
		return
	}
	defer closeFunc()

	provider, err := providers.NewProvider(providerName)
	if err != nil {
		utils.RespondError(w, record, "invalid_model_provider")
		return
	}

	var res providers.TranscriptionResult
	res.Words = make([]providers.Word, 0)
	for i, r := range files {
		resTmp, err := provider.Transcribe(ctx, r, "mpeg")
		if err != nil {
			fmt.Printf("%v\n", err)
			utils.RespondError(w, record, "transcription_error")
			return
		}
		totalStr += " " + resTmp.Text
		res.Text += " " + resTmp.Text
		res.Words = append(res.Words, resTmp.Words...)
		fmt.Printf("Transcription %v/%v\n", i+1, len(files))
	}

	res.Text = strings.Trim(totalStr, " \t\n")
	db.LogRequestsCredits(
		r.Context().Value(utils.ContextKeyEventID).(string),
		userID, "whisper", duration*1000, 0, 0, "transcription")

	response, _ := json.Marshal(&res)
	record(string(response))

	_ = json.NewEncoder(w).Encode(res)
}
