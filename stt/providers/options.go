package providers

import (
	"context"
	"errors"
	"io"
)

type Word struct {
	Word              string  `json:"word"`
	PunctuatedWord    string  `json:"punctuated_word"`
	Start             float64 `json:"start"`
	End               float64 `json:"end"`
	Confidence        float64 `json:"confidence"`
	Speaker           *int    `json:"speaker"`
	SpeakerConfidence float64 `json:"speaker_confidence"`
}

type TranscriptionResult struct {
	Text  string `json:"text"`
	Words []Word `json:"words",omitempty`
}

type Provider interface {
	Transcribe(context.Context, io.Reader, string) (*TranscriptionResult, error)
}

func NewProvider(provider string) (Provider, error) {
	if provider == "whisper" || provider == "openai" || provider == "" {
		return WhisperProvider{}, nil
	} else if provider == "deepgram" {
		return DeepgramProvider{}, nil
	}

	return nil, errors.New("invalid_provider")
}
