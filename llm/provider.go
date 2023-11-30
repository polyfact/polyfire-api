package llm

import (
	"context"
	"errors"
	"fmt"

	"github.com/polyfire/api/db"
	"github.com/polyfire/api/llm/providers"
	"github.com/polyfire/api/llm/providers/options"
	"github.com/polyfire/api/utils"
	"github.com/tmc/langchaingo/llms/cohere"
)

var ErrUnknownModel = errors.New("Unknown model")

type Provider interface {
	Name() string
	ProviderModel() (string, string)
	Generate(prompt string, c options.ProviderCallback, opts *options.ProviderOptions) chan options.Result
	DoesFollowRateLimit() bool
}

func getAvailableModels(model string) (string, string) {
	switch model {
	case "cheap":
		return "llama", "llama2"
	case "regular":
		return "openai", "gpt-3.5-turbo"
	case "best":
		return "openai", "gpt-4"
	case "uncensored":
		return "replicate", "wizard-mega-13b-awq"
	case "gpt-3.5-turbo":
		return "openai", "gpt-3.5-turbo"
	case "gpt-3.5-turbo-16k":
		return "openai", "gpt-3.5-turbo-16k"
	case "gpt-4":
		return "openai", "gpt-4"
	case "gpt-4-32k":
		return "openai", "gpt-4-32k"
	case "cohere":
		return "cohere", "cohere_command"
	case "llama-2-70b-chat":
		return "replicate", "llama-2-70b-chat"
	case "replit-code-v1-3b":
		return "replicate", "replit-code-v1-3b"
	case "wizard-mega-13b-awq":
		return "replicate", "wizard-mega-13b-awq"
	case "airoboros-llama-2-70b":
		return "replicate", "airoboros-llama-2-70b"
	case "":
		return "openai", "gpt-3.5-turbo"
	}
	return "", ""
}

func getModelWithAliases(modelAlias string, projectID string) (string, string) {
	provider, model := getAvailableModels(modelAlias)

	if model == "" {
		newModel, err := db.GetModelByAliasAndProjectID(modelAlias, projectID, "completion")
		if err != nil {
			return "", ""
		}

		model = newModel.Model
		provider = newModel.Provider
	}

	return provider, model
}

func NewProvider(ctx context.Context, modelInput string) (Provider, error) {
	projectID, _ := ctx.Value(utils.ContextKeyProjectID).(string)
	fmt.Println("Project ID: ", projectID)

	provider, model := getModelWithAliases(modelInput, projectID)

	fmt.Println("Provider: ", provider)

	switch provider {
	case "openai":
		fmt.Println("Using OpenAI")
		llm := providers.NewOpenAIStreamProvider(ctx, model)

		return llm, nil
	case "cohere":
		fmt.Println("Using Cohere")
		llm, err := cohere.New()
		if err != nil {
			return nil, err
		}
		return providers.LangchainProvider{Model: llm, ModelName: "cohere_command"}, nil
	case "llama":
		return providers.LLaMaProvider{
			Model: model,
		}, nil
	case "replicate":
		fmt.Println("Using Replicate")
		llm := providers.NewReplicateProvider(ctx, model)
		return llm, nil
	default:
		return nil, ErrUnknownModel
	}
}
