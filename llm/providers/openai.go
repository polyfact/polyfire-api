package providers

import (
	"context"
	"errors"
	"fmt"
	"math"
	"io"
	"log"
	"os"

	goOpenai "github.com/sashabaranov/go-openai"
	"github.com/tmc/langchaingo/llms"
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

func (m OpenAIStreamProvider) Generate(task string, c *func(string, int, int), opts *ProviderOptions) chan Result {
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

			tokenUsage.Input += llms.CountTokens(m.Model, task)

			totalOutput := 0

			for {
				completion, err := stream.Recv()

				if errors.Is(err, io.EOF) || err != nil {
					break
				}

				tokenUsage.Output = llms.CountTokens(m.Model, completion.Choices[0].Delta.Content)
				totalOutput += tokenUsage.Output

				result := Result{Result: completion.Choices[0].Delta.Content, TokenUsage: tokenUsage}

				chan_res <- result
			}

			if c != nil {
				(*c)(m.Model, tokenUsage.Input, totalOutput)
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
