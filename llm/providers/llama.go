package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/polyfire/api/llm/providers/options"
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

func (m LLaMaProvider) Generate(task string, c options.ProviderCallback, opts *options.ProviderOptions) chan options.Result {
	chanRes := make(chan options.Result)

	go func() {
		defer close(chanRes)
		tokenUsage := options.TokenUsage{Input: 0, Output: 0}
		body := LLaMaInputBody{Prompt: task, Model: m.Model}
		if opts != nil && opts.Temperature != nil {
			body.Temperature = opts.Temperature
		}
		fmt.Println(body)
		input, err := json.Marshal(body)
		tokenUsage.Input += tokens.CountTokens(task)
		if err != nil {
			chanRes <- options.Result{Err: "generation_error"}
			return
		}
		reqBody := string(input)
		fmt.Println(reqBody)
		resp, err := http.Post(os.Getenv("LLAMA_URL"), "application/json", strings.NewReader(reqBody))
		if err != nil {
			chanRes <- options.Result{Err: "generation_error"}
			return
		}
		defer resp.Body.Close()
		p := make([]byte, 128)
		totalOutput := 0
		totalCompletion := ""
		for {
			nb, err := resp.Body.Read(p)
			if errors.Is(err, io.EOF) || err != nil {
				break
			}
			tokenUsage.Output = tokens.CountTokens(string(p[:nb]))
			totalOutput += tokenUsage.Output
			totalCompletion += string(p[:nb])
			chanRes <- options.Result{Result: string(p[:nb]), TokenUsage: tokenUsage}
		}
		if c != nil {
			(*c)("llama", m.Model, tokenUsage.Input, totalOutput, totalCompletion, nil)
		}
	}()

	return chanRes
}

func (m LLaMaProvider) Name() string {
	return "llama"
}

func (m LLaMaProvider) ProviderModel() (string, string) {
	return "llama", m.Model
}

func (m LLaMaProvider) DoesFollowRateLimit() bool {
	return false
}
