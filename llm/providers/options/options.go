package options

import (
	"encoding/json"

	"github.com/polyfire/api/db"
	"github.com/polyfire/api/utils"
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
	Err        string           `json:"error,omitempty"`
}

type ProviderCallback *func(string, string, int, int, string, *int)

type jsonableResult struct {
	Result     string           `json:"result"`
	TokenUsage TokenUsage       `json:"token_usage"`
	Resources  []db.MatchResult `json:"ressources,omitempty"`
	Error      *utils.APIError  `json:"error,omitempty"`
}

func (r Result) JSON() ([]byte, error) {
	var apiError *utils.APIError

	if r.Err != "" {
		errorMessage, exists := utils.ErrorMessages[r.Err]

		if !exists {
			errorMessage = utils.ErrorMessages["unknown_error"]
		}
		apiError = &errorMessage
	}

	bytes, err := json.Marshal(jsonableResult{
		Result:     r.Result,
		TokenUsage: r.TokenUsage,
		Resources:  r.Resources,
		Error:      apiError,
	})
	if err != nil {
		return []byte{}, err
	}

	return bytes, nil
}
