package llm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/polyfact/api/db"
	goOpenai "github.com/sashabaranov/go-openai"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/cohere"
	"github.com/tmc/langchaingo/schema"
)

type TokenUsage struct {
	Input  int `json:"input"`
	Output int `json:"output"`
}

type Result struct {
	Result     string           `json:"result"`
	TokenUsage TokenUsage       `json:"token_usage"`
	Ressources []db.MatchResult `json:"ressources,omitempty"`
	Err        error
}

var ErrUnknownModel = errors.New("Unknown model")

type ProviderOptions struct {
	StopWords *[]string
}

type Provider interface {
	Generate(prompt string, c *func(string, int, int), opts *ProviderOptions) chan Result
}

type LLMProvider struct {
	model interface{ GetNumTokens(string) int }
}

func NewProvider(model string) (Provider, error) {
	switch model {
	case "openai":
		fmt.Println("Using OpenAI")
		llm := NewOpenAIStreamProvider()
		return llm, nil
	case "cohere":
		fmt.Println("Using Cohere")
		llm, err := cohere.New()
		if err != nil {
			return nil, err
		}
		return LLMProvider{model: llm}, nil
	default:
		return nil, ErrUnknownModel
	}
}

func (m LLMProvider) Call(prompt string, opts *ProviderOptions) (string, error) {
	ctx := context.Background()
	var result string
	var err error

	if opts == nil {
		opts = &ProviderOptions{}
	}

	options := llms.CallOptions{}

	if opts.StopWords != nil {
		options.StopWords = *opts.StopWords
	}

	if llm, ok := m.model.(llms.LLM); ok {
		result, err = llm.Call(ctx, prompt, llms.WithOptions(options))
	} else if chat, ok := m.model.(llms.ChatLLM); ok {
		result, err = chat.Call(ctx, []schema.ChatMessage{
			schema.HumanChatMessage{Text: prompt},
		}, llms.WithOptions(options))
	} else {
		return "", errors.New("Model is neither LLM nor Chat")
	}

	if err != nil {
		return "", err
	}

	return result, nil
}

func (m LLMProvider) Generate(task string, c *func(string, int, int), opts *ProviderOptions) chan Result {
	chan_res := make(chan Result)

	go func(chan_res chan Result) {
		defer close(chan_res)
		tokenUsage := TokenUsage{Input: 0, Output: 0}
		for i := 0; i < 5; i++ {
			log.Printf("Trying generation %d/5\n", i+1)

			input_prompt := task
			completion, err := m.Call(input_prompt, opts)
			if err != nil {
				fmt.Printf("%v\n", err)
				continue
			}

			if c != nil {
				(*c)(os.Getenv("OPENAI_MODEL"), m.model.GetNumTokens(input_prompt), m.model.GetNumTokens(completion))
			}

			tokenUsage.Input += m.model.GetNumTokens(input_prompt)
			tokenUsage.Output += m.model.GetNumTokens(completion)

			result := Result{Result: completion, TokenUsage: tokenUsage}

			chan_res <- result
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
	}(chan_res)

	return chan_res
}

type OpenAIStreamProvider struct {
	client goOpenai.Client
}

func NewOpenAIStreamProvider() OpenAIStreamProvider {
	config := goOpenai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	config.OrgID = os.Getenv("OPENAI_ORGANIZATION")
	return OpenAIStreamProvider{
		client: *goOpenai.NewClientWithConfig(config),
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

			req := goOpenai.ChatCompletionRequest{
				Model: goOpenai.GPT3Dot5Turbo,
				Messages: []goOpenai.ChatCompletionMessage{
					{
						Role:    goOpenai.ChatMessageRoleUser,
						Content: task,
					},
				},
				Stream: true,
			}
			stream, err := m.client.CreateChatCompletionStream(ctx, req)
			if err != nil {
				fmt.Printf("%v\n", err)
				continue
			}

			tokenUsage.Input += llms.CountTokens("gpt-3.5-turbo", task)

			for {
				completion, err := stream.Recv()

				if errors.Is(err, io.EOF) || err != nil {
					break
				}

				tokenUsage.Output += llms.CountTokens("gpt-3.5-turbo", completion.Choices[0].Delta.Content)

				result := Result{Result: completion.Choices[0].Delta.Content, TokenUsage: tokenUsage}

				chan_res <- result
			}

			if c != nil {
				(*c)(os.Getenv("OPENAI_MODEL"), tokenUsage.Input, tokenUsage.Output)
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
