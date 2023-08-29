package transcription

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/google/uuid"
	router "github.com/julienschmidt/httprouter"
	supa "github.com/nedpals/supabase-go"

	stt "github.com/polyfact/api/stt"
	"github.com/polyfact/api/utils"
)

func SplitFile(file io.Reader) ([]io.Reader, func(), error) {
	id := "split_transcribe-" + uuid.New().String()
	os.Mkdir("/tmp/"+id, 0700)
	close_func := func() {
		os.RemoveAll("/tmp/" + id)
	}
	fmt.Println(id)
	f, err := os.Create("/tmp/" + id + "/audio-file")
	if err != nil {
		return nil, nil, err
	}
	io.Copy(f, file)
	exec.Command("ffmpeg", "-i", "/tmp/"+id+"/audio-file", "/tmp/"+id+"/audio-file.ts").Run()
	os.Remove("/tmp/" + id + "/audio-file")
	split_cmd := exec.Command("split", "-b", "20971400", "/tmp/"+id+"/audio-file.ts")
	split_cmd.Dir = "/tmp/" + id
	split_cmd.Run()
	os.Remove("/tmp/" + id + "/audio-file.ts")
	files, err := ioutil.ReadDir("/tmp/" + id)
	if err != nil {
		return nil, nil, err
	}

	var res []io.Reader = make([]io.Reader, 0)
	for _, file := range files {
		if file.Name() != "audio-file" && file.Name() != "audio-file.mpeg" {
			exec.Command("ffmpeg", "-i", "/tmp/"+id+"/"+file.Name(), "/tmp/"+id+"/"+file.Name()+".mp3").Run()
			os.Remove("/tmp/" + id + "/" + file.Name())
			audio_part_r, err := os.Open("/tmp/" + id + "/" + file.Name() + ".mp3")
			if err != nil {
				return nil, nil, err
			}
			res = append(res, audio_part_r)
			fmt.Println("adding:", file.Name())
		}
	}

	return res, close_func, nil
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
	content_type := r.Header.Get("Content-Type")
	var file_size int
	var file_buf_reader io.Reader


	if content_type == "application/json" {
		var input TranscribeRequestBody

		err := json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			utils.RespondError(w, "invalid_json")
			return
		}

		b, err := DownloadFromBucket("audio_transcribes", input.FilePath)
		if err != nil {
			fmt.Println(err)
			utils.RespondError(w, "read_error")
			return
		}

		file_size = len(b)
		file_buf_reader = bytes.NewReader(b)
	} else {
		_, p, _ := mime.ParseMediaType(content_type)
		boundary := p["boundary"]
		reader := multipart.NewReader(r.Body, boundary)
		part, err := reader.NextPart()
		if err == io.EOF {
			utils.RespondError(w, "missing_content")
			return
		}
		if err != nil {
		utils.RespondError(w, "read_error", err.Error())
			return
		}
		file_buf_reader = bufio.NewReader(part)

		file_size, err = strconv.Atoi(r.Header.Get("Content-Length"))
		if err != nil {
		utils.RespondError(w, "read_error", err.Error())
			return
		}

	}

	total_str := ""
	if file_size > 25000000 {
		// The format doesn't seem to really matter
		files, close_func, err := SplitFile(file_buf_reader)
		if err != nil {
			utils.RespondError(w, "splitting_error")
		}
		defer close_func()
		for i, r := range files {
			res, err := stt.Transcribe(r, "mpeg")
			if err != nil {
				fmt.Printf("%v\n", err)
				utils.RespondError(w, "transcription_error")
				return
			}
			total_str += " " + res.Text
			fmt.Printf("Transcription %v/%v\n", i+1, len(files))
		}
	} else {
		res, err := stt.Transcribe(file_buf_reader, "mpeg")
		if err != nil {
			fmt.Printf("%v\n", err)
			utils.RespondError(w, "transcription_error")
			return
		}
		total_str = res.Text
	}

	res := Result{Text: total_str}

	json.NewEncoder(w).Encode(res)
}
