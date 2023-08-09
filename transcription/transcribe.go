package transcription

import (
	"bufio"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"

	router "github.com/julienschmidt/httprouter"
	stt "github.com/polyfact/api/stt"
	"github.com/polyfact/api/utils"
)

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
		utils.RespondError(w, "read_error")
		return
	}
	file_buf_reader := bufio.NewReader(part)

	// The format doesn't seem to really matter
	res, err := stt.Transcribe(file_buf_reader, "mp3")
	if err != nil {
		utils.RespondError(w, "transcription_error")
		return
	}

	json.NewEncoder(w).Encode(res)
}
