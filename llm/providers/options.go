package providers

import (
	"github.com/polyfact/api/db"
)

type ProviderOptions struct {
	StopWords   *[]string
	Temperature *float32
}

type TokenUsage struct {
	Input  int `json:"input"`
	Output int `json:"output"`
}

type Result struct {
	Result     string           `json:"result"`
	TokenUsage TokenUsage       `json:"token_usage"`
	Resources  []db.MatchResult `json:"ressources,omitempty"`
	Err        error
}

type ProviderCallback *func(string, string, int, int, string)
