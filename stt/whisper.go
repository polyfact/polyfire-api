package stt

import (
	"context"
	"io"
	"time"

	openai "github.com/rakyll/openai-go"
	audio "github.com/rakyll/openai-go/audio"
)

func Transcribe(reader io.Reader, format string) (*audio.CreateTranscriptionResponse, error) {
	session := openai.NewSession("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")

	(*((*session).HTTPClient)).Timeout = 600 * time.Second
	(*session).OrganizationID = "org-wEJEnFznbktHKebSaZjm6Vx8"

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
