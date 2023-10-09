package llm

import (
	"context"
	"errors"
	"fmt"

	"github.com/polyfire/api/llm/providers"
	"github.com/tmc/langchaingo/llms/cohere"
)

var ErrUnknownModel = errors.New("Unknown model")

type Provider interface {
	Name() string
	Generate(prompt string, c providers.ProviderCallback, opts *providers.ProviderOptions) chan providers.Result
	UserAllowed(user_id string) bool
	DoesFollowRateLimit() bool
}

func defaultModel(model string) (string, string) {
	switch model {
	case "cheap":
		return "llama", "llama2"
	case "regular":
		return "openai", "gpt-3.5-turbo"
	case "best":
		return "openai", "gpt-4"
	case "gpt-3.5-turbo":
		return "openai", "gpt-3.5-turbo"
	case "gpt-3.5-turbo-16k":
		return "openai", "gpt-3.5-turbo-16k"
	case "gpt-4":
		return "openai", "gpt-4"
	case "cohere":
		return "cohere", "cohere_command"
	case "codellama":
		return "llama", "codellama"
	case "llama":
		return "llama", "llama"
	case "llama2":
		return "llama", "llama2"
	case "llama-2-70b-chat":
		return "replicate", "llama-2-70b-chat"
	case "replit-code-v1-3b":
		return "replicate", "replit-code-v1-3b"
	}
	return "", ""
}

func NewProvider(ctx context.Context, provider string, model *string) (Provider, error) {
	if provider == "" && model == nil {
		provider = "openai"
	}

	if provider == "" && model != nil {
		var newModel string
		provider, newModel = defaultModel(*model)
		model = &newModel
	}

	switch provider {
	case "openai":
		fmt.Println("Using OpenAI")
		var m string

		if model == nil {
			m = "gpt-3.5-turbo"
		} else {
			m = *model
		}

		if m != "gpt-3.5-turbo" && m != "gpt-3.5-turbo-16k" && m != "gpt-4" {
			return nil, ErrUnknownModel
		}

		llm := providers.NewOpenAIStreamProvider(ctx, m)

		return llm, nil
	case "cohere":
		fmt.Println("Using Cohere")
		llm, err := cohere.New()
		if err != nil {
			return nil, err
		}
		return providers.LangchainProvider{Model: llm, ModelName: "cohere_command"}, nil
	case "llama":
		fmt.Println("Using LLama")
		var m string

		if model == nil {
			m = "llama"
		} else {
			m = *model
		}

		if m != "llama" && m != "llama2" && m != "codellama" {
			return nil, ErrUnknownModel
		}

		return providers.LLaMaProvider{
			Model: m,
		}, nil
	case "replicate":
		fmt.Println("Using Replicate")
		var m string
		if model == nil {
			m = "llama-2-70b-chat"
		} else {
			m = *model
		}

		if m != "llama-2-70b-chat" && m != "replit-code-v1-3b" {
			return nil, ErrUnknownModel
		}

		llm := providers.NewReplicateProvider(ctx, m)
		return llm, nil
	default:
		return nil, ErrUnknownModel
	}
}
