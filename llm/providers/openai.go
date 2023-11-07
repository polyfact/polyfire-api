package providers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	db "github.com/polyfire/api/db"
	"github.com/polyfire/api/llm/providers/options"
	tokens "github.com/polyfire/api/tokens"
	utils "github.com/polyfire/api/utils"
	goOpenai "github.com/sashabaranov/go-openai"
)

type OpenAIStreamProvider struct {
	client        goOpenai.Client
	Model         string
	IsCustomToken bool
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
	return OpenAIStreamProvider{
		client:        *goOpenai.NewClientWithConfig(config),
		Model:         model,
		IsCustomToken: isCustomToken,
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

		fmt.Printf("%v\n", req.Temperature)

		stream, err := m.client.CreateChatCompletionStream(ctx, req)
		if err != nil {
			if strings.Contains(err.Error(), "Incorrect API key provided") && m.IsCustomToken {
				chanRes <- options.Result{Err: "openai_invalid_api_key"}
			} else {
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

			fmt.Printf("%v %v\n", completion.Choices[0].Delta.Content, tokenUsage.Output)

			totalOutput += tokenUsage.Output

			result := options.Result{Result: completion.Choices[0].Delta.Content, TokenUsage: tokenUsage}

			totalCompletion += completion.Choices[0].Delta.Content

			chanRes <- result
		}
		if c != nil {
			(*c)("openai", m.Model, tokenUsage.Input, totalOutput, totalCompletion, nil)
		}
	}()

	return chanRes
}

func (m OpenAIStreamProvider) UserAllowed(userID string) bool {
	if m.Model == "gpt-3.5-turbo" || m.Model == "gpt-3.5-turbo-16k" {
		return true
	}

	res, _ := db.ProjectIsPremium(userID)
	return res
}

func (m OpenAIStreamProvider) Name() string {
	return "openai"
}

func (m OpenAIStreamProvider) ProviderModel() (string, string) {
	return "openai", m.Model
}

func (m OpenAIStreamProvider) DoesFollowRateLimit() bool {
	return !m.IsCustomToken
}
