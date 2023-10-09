package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	tokens "github.com/polyfire/api/tokens"
)

type LLaMaInputBody struct {
	Prompt      string   `json:"prompt"`
	Model       string   `json:"model"`
	Temperature *float32 `json:"temperature"`
}

type LLaMaProvider struct {
	Model string
}

func (m LLaMaProvider) Generate(task string, c ProviderCallback, opts *ProviderOptions) chan Result {
	chan_res := make(chan Result)

	go func() {
		defer close(chan_res)
		tokenUsage := TokenUsage{Input: 0, Output: 0}
		body := LLaMaInputBody{Prompt: task, Model: m.Model}
		if opts != nil && opts.Temperature != nil {
			body.Temperature = opts.Temperature
		}
		fmt.Println(body)
		input, err := json.Marshal(body)
		tokenUsage.Input += tokens.CountTokens("gpt-2", task)
		if err != nil {
			chan_res <- Result{Err: "generation_error"}
			return
		}
		reqBody := string(input)
		fmt.Println(reqBody)
		resp, err := http.Post(os.Getenv("LLAMA_URL"), "application/json", strings.NewReader(reqBody))
		if err != nil {
			chan_res <- Result{Err: "generation_error"}
			return
		}
		defer resp.Body.Close()
		var p []byte = make([]byte, 128)
		totalOutput := 0
		totalCompletion := ""
		for {
			nb, err := resp.Body.Read(p)
			if errors.Is(err, io.EOF) || err != nil {
				break
			}
			tokenUsage.Output = tokens.CountTokens("gpt-2", string(p[:nb]))
			totalOutput += tokenUsage.Output
			totalCompletion += string(p[:nb])
			chan_res <- Result{Result: string(p[:nb]), TokenUsage: tokenUsage}
		}
		if c != nil {
			(*c)("llama", m.Model, tokenUsage.Input, totalOutput, totalCompletion, nil)
		}
	}()

	return chan_res
}

func (m LLaMaProvider) UserAllowed(_user_id string) bool {
	return true
}

func (m LLaMaProvider) Name() string {
	return "llama"
}

func (m LLaMaProvider) DoesFollowRateLimit() bool {
	return false
}
