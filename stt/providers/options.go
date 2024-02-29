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

type TranscriptionInputOptions struct {
	Format       string
	Language     *string
	OutputFormat *string
}

type DialogueElement struct {
	Speaker int     `json:"speaker"`
	Text    string  `json:"text"`
	Start   float64 `json:"start"`
	End     float64 `json:"end"`
}

type TranscriptionResult struct {
	Text     string            `json:"text"`
	Words    []Word            `json:"words,omitempty"`
	Dialogue []DialogueElement `json:"dialogue,omitempty"`
}

type Provider interface {
	Transcribe(context.Context, io.Reader, TranscriptionInputOptions) (*TranscriptionResult, error)
}

func NewProvider(provider string) (Provider, error) {
	if provider == "whisper" || provider == "openai" || provider == "" {
		return WhisperProvider{}, nil
	} else if provider == "deepgram" {
		return DeepgramProvider{}, nil
	} else if provider == "assemblyai" {
		return AssemblyAIProvider{}, nil
	}

	return nil, errors.New("invalid_provider")
}
