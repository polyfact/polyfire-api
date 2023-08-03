package providers

import (
	"github.com/polyfact/api/db"
)

type ProviderOptions struct {
	StopWords *[]string
}

type TokenUsage struct {
	Input  int `json:"input"`
	Output int `json:"output"`
}

type Result struct {
	Result     string           `json:"result"`
	TokenUsage TokenUsage       `json:"token_usage"`
	Ressources []db.MatchResult `json:"ressources,omitempty"`
	Err        error
}
