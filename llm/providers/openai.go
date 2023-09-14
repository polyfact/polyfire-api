package providers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"

	db "github.com/polyfact/api/db"
	tokens "github.com/polyfact/api/tokens"
	goOpenai "github.com/sashabaranov/go-openai"
)

type OpenAIStreamProvider struct {
	client goOpenai.Client
	Model  string
}

func NewOpenAIStreamProvider(model string) OpenAIStreamProvider {
	config := goOpenai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	config.OrgID = os.Getenv("OPENAI_ORGANIZATION")
	return OpenAIStreamProvider{
		client: *goOpenai.NewClientWithConfig(config),
		Model:  model,
	}
}

func (m OpenAIStreamProvider) Generate(task string, c *func(string, string, int, int), opts *ProviderOptions) chan Result {
	chan_res := make(chan Result)

	go func() {
		defer close(chan_res)
		tokenUsage := TokenUsage{Input: 0, Output: 0}
		for i := 0; i < 5; i++ {
			log.Printf("Trying generation %d/5\n", i+1)

			ctx := context.Background()

			if opts == nil {
				opts = &ProviderOptions{}
			}

			req := goOpenai.ChatCompletionRequest{
				Model: m.Model,
				Messages: []goOpenai.ChatCompletionMessage{
					{
						Role:    goOpenai.ChatMessageRoleUser,
						Content: task,
					},
				},
				Stream: true,
			}

			if opts.StopWords != nil {
				req.Stop = *opts.StopWords
			}
			if opts.Temperature != nil {
				if *opts.Temperature == 0.0 {
					var nearly_zero float32 = math.SmallestNonzeroFloat32
					req.Temperature = nearly_zero // We need to do that bc openai-go omitempty on 0.0
				} else {
					req.Temperature = *opts.Temperature
				}
			}

			fmt.Printf("%v\n", req.Temperature)

			stream, err := m.client.CreateChatCompletionStream(ctx, req)
			if err != nil {
				fmt.Printf("%v\n", err)
				continue
			}

			tokenUsage.Input += tokens.CountTokens(m.Model, task)

			totalOutput := 0

			for {
				completion, err := stream.Recv()

				if errors.Is(err, io.EOF) || err != nil {
					break
				}

				tokenUsage.Output = tokens.CountTokens(m.Model, completion.Choices[0].Delta.Content)

				fmt.Printf("%v %v\n", completion.Choices[0].Delta.Content, tokenUsage.Output)

				totalOutput += tokenUsage.Output

				result := Result{Result: completion.Choices[0].Delta.Content, TokenUsage: tokenUsage}

				chan_res <- result
			}

			if c != nil {
				(*c)("openai", m.Model, tokenUsage.Input, totalOutput)
			}
			return
		}
		chan_res <- Result{
			Result:     "{\"error\":\"generation_failed\"}",
			TokenUsage: tokenUsage,
			Err: errors.New(
				"Generation failed after 5 retries",
			),
		}
		return
	}()

	return chan_res
}

func (m OpenAIStreamProvider) UserAllowed(user_id string) bool {
	if m.Model == "gpt-3.5-turbo" || m.Model == "gpt-3.5-turbo-16k" {
		return true
	}

	res, _ := db.ProjectIsPremium(user_id)
	return res
}
