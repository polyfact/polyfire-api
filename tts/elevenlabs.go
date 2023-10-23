package tts

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/haguro/elevenlabs-go"
	router "github.com/julienschmidt/httprouter"

	"github.com/polyfire/api/db"
	"github.com/polyfire/api/utils"
)

var (
	RateLimitReached        = errors.New("Rate limit reached")
	ProjectRateLimitReached = errors.New("Project rate limit reached")
	UnknownError            = errors.New("Unknown error")
)

func TextToSpeech(ctx context.Context, w io.Writer, text string, voiceID string) error {
	customToken, ok := ctx.Value(utils.ContextKeyElevenlabsToken).(string)
	if !ok {
		customToken = os.Getenv("ELEVENLABS_API_KEY")
		rateLimitStatus := ctx.Value(utils.ContextKeyRateLimitStatus)
		if rateLimitStatus != db.RateLimitStatusOk {
			return RateLimitReached
		}
	}

	client := elevenlabs.NewClient(context.Background(), customToken, 30*time.Second)

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
	userId := r.Context().Value(utils.ContextKeyUserID).(string)

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

	db.LogRequestsCredits(userId, "elevenlabs", "elevenlabs", len(reqBody.Text)*3000, len(reqBody.Text), 0, "tts")
	err = TextToSpeech(r.Context(), w, reqBody.Text, voice.ProviderVoiceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
