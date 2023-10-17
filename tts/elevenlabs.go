package tts

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/haguro/elevenlabs-go"
	router "github.com/julienschmidt/httprouter"

	"github.com/polyfire/api/db"
)

func TextToSpeech(w io.Writer, text string, voiceID string) error {
	client := elevenlabs.NewClient(context.Background(), os.Getenv("ELEVENLABS_API_KEY"), 30*time.Second)

	ttsReq := elevenlabs.TextToSpeechRequest{
		Text:    text,
		ModelID: "eleven_multilingual_v2",
	}

	return client.TextToSpeechStream(w, voiceID, ttsReq)
}

type TTSRequestBody struct {
	Text  string  `json:"text"`
	Voice *string `json:"voice"`
}

func TTSHandler(w http.ResponseWriter, r *http.Request, _ router.Params) {
	var reqBody TTSRequestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var voiceSlug string
	if (reqBody.Voice == nil) || (*reqBody.Voice == "") {
		voiceSlug = "default"
	} else {
		voiceSlug = *reqBody.Voice
	}

	voice, err := db.GetTTSVoice(voiceSlug)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "audio/mp3")

	err = TextToSpeech(w, reqBody.Text, voice.ProviderVoiceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
