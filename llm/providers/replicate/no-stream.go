package providers

import (
	"encoding/json"
	"fmt"
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
	chanRes := make(chan options.Result)

	go func() {
		defer close(chanRes)

		replicateStartTime := time.Now()
		replicateAfterBootTime := time.Now()

		startResponse, errorCode := m.ReplicateStart(task, opts, false)
		if errorCode != "" {
			chanRes <- options.Result{Err: errorCode}
			return
		}

		var completion string
		tokenUsage := options.TokenUsage{}

		tokenUsage.Input = tokens.CountTokens(task)
		coldBootDetected := false

		for {
			respBody, err := m.SendRequest(startResponse.URLs.Get)
			if err != nil {
				fmt.Println(err)
				chanRes <- options.Result{Err: "generation_error"}
				return
			}

			var output ReplicatePredictionOutput
			err = json.NewDecoder(respBody).Decode(&output)
			if err != nil || output.Status == "error" {
				fmt.Println(err)
				chanRes <- options.Result{Err: "generation_error"}
				return
			}

			if output.Status == "starting" {
				replicateAfterBootTime = time.Now()
				if time.Since(replicateStartTime) > 10*time.Second && !coldBootDetected {
					fmt.Println("cold boot detected")
					chanRes <- options.Result{Warnings: []string{"The model is taking longer than usual to start up. It's probably due to a cold boot on replicate's side. It will respond enventually but it can take some time."}}
					coldBootDetected = true
				}
			}

			if output.Status == "succeeded" {
				completion = output.Output
				tokenUsage.Output = tokens.CountTokens(completion)
				chanRes <- options.Result{Result: output.Output, TokenUsage: tokenUsage}
				break
			}

			time.Sleep(1 * time.Second)
		}

		replicateEndTime := time.Now()
		duration := replicateEndTime.Sub(replicateAfterBootTime)

		if c != nil {
			credits := int(duration.Seconds()*m.CreditsPerSecond) + 1
			(*c)("replicate", m.Model, tokenUsage.Input, tokenUsage.Output, completion, &credits)
		}
	}()

	return chanRes
}
