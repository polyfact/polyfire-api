package llm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/cohere"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
)

type TokenUsage struct {
	Input  int `json:"input"`
	Output int `json:"output"`
}

type Result struct {
	Result     string     `json:"result"`
	TokenUsage TokenUsage `json:"token_usage"`
	Err        error
}

var ErrUnknownModel = errors.New("Unknown model")

type ProviderOptions struct {
	StopWords *[]string
}

type Provider interface {
	Generate(prompt string, c *func(string, int, int), opts *ProviderOptions) chan string
}

type LLMProvider struct {
	model interface{ GetNumTokens(string) int }
}

func NewLLMProvider(model string) (*LLMProvider, error) {
	switch model {
	case "openai":
		fmt.Println("Using OpenAI")
		llm, err := openai.NewChat()
		if err != nil {
			return nil, err
		}
		return &LLMProvider{model: llm}, nil
	case "cohere":
		fmt.Println("Using Cohere")
		llm, err := cohere.New()
		if err != nil {
			return nil, err
		}
		return &LLMProvider{model: llm}, nil
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

	go func() {
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
	}()

	return chan_res
}
