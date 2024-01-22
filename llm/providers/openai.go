package providers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"

	"github.com/polyfire/api/llm/providers/options"
	tokens "github.com/polyfire/api/tokens"
	utils "github.com/polyfire/api/utils"
	goOpenai "github.com/sashabaranov/go-openai"
)

type OpenAIStreamProvider struct {
	Client        goOpenai.Client
	Model         string
	IsCustomToken bool
	Provider      string
}

func NewOpenAIStreamProvider(ctx context.Context, model string) OpenAIStreamProvider {
	var config goOpenai.ClientConfig
	var isCustomToken bool

	customToken, ok := ctx.Value(utils.ContextKeyOpenAIToken).(string)
	if ok {
		config = goOpenai.DefaultConfig(customToken)
		customOrg, ok := ctx.Value(utils.ContextKeyOpenAIOrg).(string)
		if ok {
			config.OrgID = customOrg
		}
		isCustomToken = true
	} else {
		config = goOpenai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
		config.OrgID = os.Getenv("OPENAI_ORGANIZATION")
		isCustomToken = false
	}

	if client, ok := ctx.Value(utils.ContextKeyHTTPClient).(*http.Client); ok {
		config.HTTPClient = client
	}

	if base, ok := ctx.Value(utils.ContextKeyOpenAIBaseURL).(string); ok {
		config.BaseURL = base
	}

	return OpenAIStreamProvider{
		Client:        *goOpenai.NewClientWithConfig(config),
		Model:         model,
		IsCustomToken: isCustomToken,
		Provider:      "openai",
	}
}

func (m OpenAIStreamProvider) Generate(
	task string,
	c options.ProviderCallback,
	opts *options.ProviderOptions,
) chan options.Result {
	chanRes := make(chan options.Result)

	go func() {
		defer close(chanRes)
		tokenUsage := options.TokenUsage{Input: 0, Output: 0}
		ctx := context.Background()

		if opts == nil {
			opts = &options.ProviderOptions{}
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
				var nearlyZero float32 = math.SmallestNonzeroFloat32
				req.Temperature = nearlyZero // We need to do that bc openai-go omitempty on 0.0
			} else {
				req.Temperature = *opts.Temperature
			}
		}

		stream, err := m.Client.CreateChatCompletionStream(ctx, req)
		if err != nil {
			if strings.Contains(err.Error(), "Incorrect API key provided") && m.IsCustomToken {
				chanRes <- options.Result{Err: "openai_invalid_api_key"}
			} else {
				fmt.Println(err)
				chanRes <- options.Result{Err: "generation_error"}
			}
			return
		}

		tokenUsage.Input += tokens.CountTokens(task)

		totalOutput := 0
		totalCompletion := ""

		for {
			completion, err := stream.Recv()

			if errors.Is(err, io.EOF) || err != nil {
				break
			}
			tokenUsage.Output = tokens.CountTokens(completion.Choices[0].Delta.Content)

			totalOutput += tokenUsage.Output

			result := options.Result{Result: completion.Choices[0].Delta.Content, TokenUsage: tokenUsage}

			totalCompletion += completion.Choices[0].Delta.Content

			chanRes <- result
		}
		if c != nil {
			(*c)(m.Provider, m.Model, tokenUsage.Input, totalOutput, totalCompletion, nil)
		}
	}()

	return chanRes
}

func (m OpenAIStreamProvider) Name() string {
	return m.Provider
}

func (m OpenAIStreamProvider) ProviderModel() (string, string) {
	return m.Provider, m.Model
}

func (m OpenAIStreamProvider) DoesFollowRateLimit() bool {
	return !m.IsCustomToken
}
