package transcription

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"

	stt "github.com/polyfact/api/stt"
)

func Transcribe(w http.ResponseWriter, r *http.Request) {
	_, p, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	boundary := p["boundary"]
	reader := multipart.NewReader(r.Body, boundary)
	part, err := reader.NextPart()
	if err == io.EOF {
		http.Error(w, "400 Bad Request", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
		return
	}
	file_buf_reader := bufio.NewReader(part)

	// The format doesn't seem to really matter
	res, err := stt.Transcribe(file_buf_reader, "mp3")
	if err != nil {
		log.Printf("%w", err)
		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
}
