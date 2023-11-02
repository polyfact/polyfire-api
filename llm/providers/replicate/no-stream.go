package providers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/polyfire/api/llm/providers/options"
	"github.com/polyfire/api/tokens"
)

type ReplicatePredictionOutput struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Output string `json:"output"`
}

func (m ReplicateProvider) NoStream(
	task string,
	c options.ProviderCallback,
	opts *options.ProviderOptions,
) chan options.Result {
	chan_res := make(chan options.Result)

	go func() {
		defer close(chan_res)

		replicateStartTime := time.Now()

		startResponse, errorCode := m.ReplicateStart(task, opts, false)
		if errorCode != "" {
			chan_res <- options.Result{Err: errorCode}
			return
		}

		var completion string
		tokenUsage := options.TokenUsage{}

		tokenUsage.Input = tokens.CountTokens("gpt-2", task)
		var coldBootDetected bool = false

		for {
			req, err := http.NewRequest("GET", startResponse.URLs.Get, nil)
			if err != nil {
				fmt.Println(err)
				chan_res <- options.Result{Err: "generation_error"}
				return
			}

			req.Header.Set("Authorization", "Token "+m.ReplicateApiKey)
			req.Header.Set("Accept", "text/event-stream")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				fmt.Println(err)
				chan_res <- options.Result{Err: "generation_error"}
				return
			}

			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println(err)
				chan_res <- options.Result{Err: "generation_error"}
				return
			}

			var output ReplicatePredictionOutput
			err = json.Unmarshal(respBody, &output)
			if err != nil {
				fmt.Println(err)
				chan_res <- options.Result{Err: "generation_error"}
				return
			}

			if output.Status == "error" {
				fmt.Println(output)
				chan_res <- options.Result{Err: "generation_error"}
				return
			}

			if output.Status == "starting" && time.Since(replicateStartTime) > 10*time.Second && !coldBootDetected {
				fmt.Println("cold boot detected")
				chan_res <- options.Result{Warnings: []string{"The model is taking longer than usual to start up. It's probably due to a cold boot on replicate's side. It will respond enventually but it can take some time."}}
				coldBootDetected = true
			}

			if output.Status == "succeeded" {
				completion = output.Output
				tokenUsage.Output = tokens.CountTokens("gpt-2", completion)
				chan_res <- options.Result{Result: output.Output, TokenUsage: tokenUsage}
				break
			}

			time.Sleep(1 * time.Second)
		}

		replicateEndTime := time.Now()
		duration := replicateEndTime.Sub(replicateStartTime)

		if c != nil {
			credits := int(duration.Seconds()*m.CreditsPerSecond) + 1
			(*c)("replicate", m.Model, tokenUsage.Input, tokenUsage.Output, completion, &credits)
		}
	}()

	return chan_res
}
