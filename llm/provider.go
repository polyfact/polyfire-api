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
}

var ErrUnknownModel = errors.New("Unknown model")

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

func (m LLMProvider) Call(prompt string, opts *llms.CallOptions) (string, error) {
	ctx := context.Background()
	var result string
	var err error

	if opts == nil {
		opts = &llms.CallOptions{}
	}

	if llm, ok := m.model.(llms.LLM); ok {
		result, err = llm.Call(ctx, prompt, llms.WithOptions(*opts))
	} else if chat, ok := m.model.(llms.ChatLLM); ok {
		result, err = chat.Call(ctx, []schema.ChatMessage{
			schema.HumanChatMessage{Text: prompt},
		}, llms.WithOptions(*opts))
	} else {
		return "", errors.New("Model is neither LLM nor Chat")
	}

	if err != nil {
		return "", err
	}

	return result, nil
}

func (m LLMProvider) Generate(task string, c *func(string, int, int), opts *llms.CallOptions) (Result, error) {
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

		return Result{Result: completion, TokenUsage: tokenUsage}, err
	}

	return Result{
			Result:     "{\"error\":\"generation_failed\"}",
			TokenUsage: tokenUsage,
		}, errors.New(
			"Generation failed after 5 retries",
		)
}
