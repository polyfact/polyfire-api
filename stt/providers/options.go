package providers

import (
	"context"
	"encoding/json"
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

type KeywordBoost struct {
	Keyword string
	Boost   float64
}

func (k *KeywordBoost) UnmarshalJSON(data []byte) error {
	var keywordWithBoost struct {
		Keyword string  `json:"keyword"`
		Boost   float64 `json:"boost"`
	}

	var keyword string

	if err := json.Unmarshal(data, &keywordWithBoost); err == nil {
		*k = KeywordBoost{
			Keyword: keywordWithBoost.Keyword,
			Boost:   keywordWithBoost.Boost,
		}
	} else if err := json.Unmarshal(data, &keyword); err == nil {
		*k = KeywordBoost{
			Keyword: keyword,
			Boost:   1,
		}
	} else {
		return errors.New("Could not unmarshal keyword object")
	}

	return nil
}

type TranscriptionInputOptions struct {
	Format       string
	Language     *string
	OutputFormat *string
	Keywords     []KeywordBoost
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
	switch provider {
	case "whisper", "openai", "":
		return WhisperProvider{}, nil
	case "deepgram":
		return DeepgramProvider{}, nil
	case "google":
		return GoogleProvider{}, nil
	case "assemblyai":
		return AssemblyAIProvider{}, nil
	default:
		return nil, errors.New("invalid_provider")
	}
}
