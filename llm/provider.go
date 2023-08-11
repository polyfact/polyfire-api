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

func NewProvider(model string) (Provider, error) {
	switch model {
	case "openai":
		fmt.Println("Using OpenAI")
		llm := providers.NewOpenAIStreamProvider()
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
		return providers.LLaMaProvider{}, nil
	default:
		return nil, ErrUnknownModel
	}
}
