package tts

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/haguro/elevenlabs-go"
	router "github.com/julienschmidt/httprouter"
)

func TextToSpeech(text string, voiceID string) ([]byte, error) {
	client := elevenlabs.NewClient(context.Background(), os.Getenv("ELEVENLABS_API_KEY"), 30*time.Second)

	ttsReq := elevenlabs.TextToSpeechRequest{
		Text:    text,
		ModelID: "eleven_multilingual_v2",
	}

	audio, err := client.TextToSpeech(voiceID, ttsReq)
	if err != nil {
		return nil, err
	}

	return audio, nil
}

type TTSRequestBody struct {
	Text string `json:"text"`
}

func TTSHandler(w http.ResponseWriter, r *http.Request, _ router.Params) {
	var reqBody TTSRequestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	audio, err := TextToSpeech(reqBody.Text, "GBv7mTt0atIp3Br8iCZE")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "audio/mp3")

	w.Write(audio)
}
