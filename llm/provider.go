package llm

import (
	"errors"
	"fmt"

	"github.com/polyfact/api/llm/providers"
	"github.com/tmc/langchaingo/llms/cohere"
)

var ErrUnknownModel = errors.New("Unknown model")

type Provider interface {
	Generate(prompt string, c *func(string, int, int), opts *providers.ProviderOptions) chan providers.Result
}

func NewProvider(provider string, model *string) (Provider, error) {
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
		llm := providers.NewOpenAIStreamProvider(m)
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
		if m != "llama" && m != "llama2" {
			return nil, ErrUnknownModel
		}
		return providers.LLaMaProvider{
			Model: m,
		}, nil
	default:
		return nil, ErrUnknownModel
	}
}
