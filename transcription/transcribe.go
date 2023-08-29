package transcription

import (
	"bufio"
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

	stt "github.com/polyfact/api/stt"
	"github.com/polyfact/api/utils"
)

func SplitFile(file io.Reader) ([]io.Reader, func()) {
	id := "split_transcribe-" + uuid.New().String()
	os.Mkdir("/tmp/"+id, 0700)
	close_func := func() {
		os.RemoveAll("/tmp/" + id)
	}
	fmt.Println(id)
	f, err := os.Create("/tmp/" + id + "/audio-file")
	if err != nil {
		panic(err)
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
		panic(err)
	}

	var res []io.Reader = make([]io.Reader, 0)
	for _, file := range files {
		if file.Name() != "audio-file" && file.Name() != "audio-file.mpeg" {
			exec.Command("ffmpeg", "-i", "/tmp/"+id+"/"+file.Name(), "/tmp/"+id+"/"+file.Name()+".mp3").Run()
			os.Remove("/tmp/" + id + "/" + file.Name())
			audio_part_r, err := os.Open("/tmp/" + id + "/" + file.Name() + ".mp3")
			if err != nil {
				panic(err)
			}
			res = append(res, audio_part_r)
			fmt.Println("adding:", file.Name())
		}
	}

	return res, close_func
}

type Result struct {
	Text string `json:"text"`
}

func Transcribe(w http.ResponseWriter, r *http.Request, _ router.Params) {
	_, p, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
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
	file_buf_reader := bufio.NewReader(part)

	total_str := ""

	content_length, err := strconv.Atoi(r.Header.Get("Content-Length"))
	if err != nil {
		utils.RespondError(w, "read_error", err.Error())
		return
	}

	if content_length > 25000000 {
		// The format doesn't seem to really matter
		files, close_func := SplitFile(file_buf_reader)
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
