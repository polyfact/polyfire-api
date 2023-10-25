package providers

import (
	"context"
	"io"
	"os"
	"time"

	utils "github.com/polyfire/api/utils"
	openai "github.com/rakyll/openai-go"
	audio "github.com/rakyll/openai-go/audio"
)

func WhisperTranscribe(ctx context.Context, reader io.Reader, format string) (*TranscriptionResult, error) {
	var session *openai.Session
	customToken, ok := ctx.Value(utils.ContextKeyOpenAIToken).(string)
	if ok {
		session = openai.NewSession(customToken)
		customOrg, ok := ctx.Value(utils.ContextKeyOpenAIOrg).(string)
		if ok {
			(*session).OrganizationID = customOrg
		}
	} else {
		session = openai.NewSession(os.Getenv("OPENAI_API_KEY"))
		(*session).OrganizationID = os.Getenv("OPENAI_ORGANIZATION")
	}

	(*((*session).HTTPClient)).Timeout = 600 * time.Second

	client := audio.NewClient(session, "whisper-1")

	params := audio.CreateTranscriptionParams{
		Audio:       reader,
		AudioFormat: format,
	}

	transcriptionCtx := context.Background()

	response, err := client.CreateTranscription(transcriptionCtx, &params)
	if err != nil {
		return nil, err
	}

	res := TranscriptionResult{
		Text: response.Text,
	}
	return &res, nil
}
