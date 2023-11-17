package tts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/haguro/elevenlabs-go"
	router "github.com/julienschmidt/httprouter"

	"github.com/polyfire/api/db"
	"github.com/polyfire/api/utils"
)

var (
	ErrRateLimitReached        = errors.New("Rate limit reached")
	ErrProjectRateLimitReached = errors.New("Project rate limit reached")
	ErrCreditsUsedUp           = errors.New("Credits Used Up")
	ErrUnknown                 = errors.New("Unknown error")
)

func TextToSpeech(ctx context.Context, w io.Writer, text string, voiceID string) error {
	customToken, ok := ctx.Value(utils.ContextKeyElevenlabsToken).(string)
	if !ok {
		customToken = os.Getenv("ELEVENLABS_API_KEY")
		rateLimitStatus := ctx.Value(utils.ContextKeyRateLimitStatus)
		if rateLimitStatus != db.RateLimitStatusOk {
			return ErrRateLimitReached
		}

		creditsStatus := ctx.Value(utils.ContextKeyCreditsStatus)

		if creditsStatus == db.CreditsStatusUsedUp {
			return ErrCreditsUsedUp
		}

	}

	client := elevenlabs.NewClient(context.Background(), customToken, 30*time.Second)

	ttsReq := elevenlabs.TextToSpeechRequest{
		Text:    text,
		ModelID: "eleven_multilingual_v2",
	}

	return client.TextToSpeechStream(w, voiceID, ttsReq)
}

type RequestBody struct {
	Text  string  `json:"text"`
	Voice *string `json:"voice"`
}

func Handler(w http.ResponseWriter, r *http.Request, _ router.Params) {
	userID := r.Context().Value(utils.ContextKeyUserID).(string)
	record := r.Context().Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)

	var reqBody RequestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		utils.RespondError(w, record, "invalid_json")
		return
	}

	var voiceSlug string
	if (reqBody.Voice == nil) || (*reqBody.Voice == "") {
		voiceSlug = "default"
	} else {
		voiceSlug = *reqBody.Voice
	}

	voice, err := db.GetTTSVoice(strings.ToLower(voiceSlug))
	if err != nil {
		utils.RespondError(w, record, "voice_not_found")
		return
	}

	w.Header().Set("Content-Type", "audio/mp3")

	db.LogRequestsCredits(
		r.Context().Value(utils.ContextKeyEventID).(string),
		userID,
		"elevenlabs",
		len(reqBody.Text)*3000,
		len(reqBody.Text),
		0,
		"tts",
	)
	err = TextToSpeech(r.Context(), w, reqBody.Text, voice.ProviderVoiceID)
	if err != nil {
		fmt.Println(err)
		utils.RespondError(w, record, "elevenlabs_error")
		return
	}

	record("[Raw audio]")
}
