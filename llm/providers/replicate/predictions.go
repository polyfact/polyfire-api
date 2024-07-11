package providers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/polyfire/api/llm/providers/options"
)

type ReplicateProvider struct {
	Model            string
	ReplicateAPIKey  string
	IsCustomAPIKey   bool
	Version          string
	CreditsPerSecond float64
}

type ReplicateInput struct {
	Prompt  string `json:"prompt"`
	Task    string `json:"task"`
	Message string `json:"message"`
	Text    string `json:"text"`

	Temperature *float32 `json:"temperature,omitempty"`
}

type ReplicateRequestBody struct {
	Version string         `json:"version"`
	Input   ReplicateInput `json:"input"`
	Stream  bool           `json:"stream"`
}

type ReplicateStartResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	URLs   struct {
		Stream string `json:"stream"`
		Get    string `json:"get"`
		Cancel string `json:"cancel"`
	} `json:"urls"`
}

type ReplicateStartErrorResponse struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
	Status int    `json:"status"`
}

func (m ReplicateProvider) ReplicateStart(
	task string,
	opts *options.ProviderOptions,
	stream bool,
) (ReplicateStartResponse, string) {
	reqBody := ReplicateRequestBody{
		Input:   ReplicateInput{},
		Version: m.Version,
		Stream:  stream,
	}

	if opts != nil && opts.Temperature != nil {
		reqBody.Input.Temperature = opts.Temperature
	}

	reqBody.Input.Task = task
	reqBody.Input.Prompt = task
	reqBody.Input.Message = task
	reqBody.Input.Text = task

	input, err := json.Marshal(reqBody)
	if err != nil {
		return ReplicateStartResponse{}, "generation_error"
	}

	req, err := http.NewRequest(
		"POST",
		"https://api.replicate.com/v1/predictions",
		strings.NewReader(string(input)),
	)
	if err != nil {
		return ReplicateStartResponse{}, "generation_error"
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+m.ReplicateAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ReplicateStartResponse{}, "generation_error"
	}

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return ReplicateStartResponse{}, "generation_error"
	}

	fmt.Printf("%v\n", string(res))

	var startResponse ReplicateStartResponse
	err = json.Unmarshal(res, &startResponse)
	if err != nil {
		var errorResponse ReplicateStartErrorResponse
		err = json.Unmarshal(res, &errorResponse)
		if err != nil {
			return ReplicateStartResponse{}, "generation_error"
		}

		if errorResponse.Title == "Unauthenticated" && m.IsCustomAPIKey {
			return ReplicateStartResponse{}, "replicate_unauthenticated"
		}

		if errorResponse.Title == "Invalid version or not permitted" && m.IsCustomAPIKey {
			return ReplicateStartResponse{}, "replicate_invalid_version_or_forbidden"
		}

		fmt.Printf("%v\n", errorResponse)

		return ReplicateStartResponse{}, "generation_error"
	}

	return startResponse, ""
}
