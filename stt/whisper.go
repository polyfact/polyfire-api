package stt

import (
	"context"
	"io"
	"os"
	"time"

	openai "github.com/rakyll/openai-go"
	audio "github.com/rakyll/openai-go/audio"
)

func Transcribe(reader io.Reader, format string) (*audio.CreateTranscriptionResponse, error) {
	session := openai.NewSession(os.Getenv("OPENAI_API_KEY"))

	(*((*session).HTTPClient)).Timeout = 600 * time.Second
	(*session).OrganizationID = os.Getenv("OPENAI_ORGANIZATION")

	client := audio.NewClient(session, "whisper-1")

	params := audio.CreateTranscriptionParams{
		Audio:       reader,
		AudioFormat: format,
	}

	ctx := context.Background()

	response, err := client.CreateTranscription(ctx, &params)
	if err != nil {
		return nil, err
	}

	return response, nil
}
