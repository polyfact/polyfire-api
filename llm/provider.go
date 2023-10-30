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
	UserAllowed(user_id string) bool
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
	}
	return "", ""
}

func getModelWithAliases(modelAlias string, projectId string) (string, string) {
	provider, model := getAvailableModels(modelAlias)

	fmt.Println("Project ID: ", projectId)
	if model == "" {
		newModel, err := db.GetModelByAliasAndProjectId(modelAlias, projectId, "completion")
		if err != nil {
			return "", ""
		}

		model = newModel.Model
		provider = newModel.Provider
	}

	return provider, model
}

func NewProvider(ctx context.Context, provider string, model *string) (Provider, error) {
	if provider == "" && model == nil {
		provider = "openai"
	}

	if provider == "" && model != nil {
		projectId, _ := ctx.Value(utils.ContextKeyProjectID).(string)

		var newModel string
		provider, newModel = getModelWithAliases(*model, projectId)
		model = &newModel
	}

	fmt.Println("Provider: ", provider)

	switch provider {
	case "openai":
		fmt.Println("Using OpenAI")
		var m string

		if model == nil {
			m = "gpt-3.5-turbo"
		} else {
			m = *model
		}

		if m != "gpt-3.5-turbo" && m != "gpt-3.5-turbo-16k" && m != "gpt-4" && m != "gpt-4-32k" {
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

		if m != "llama-2-70b-chat" && m != "replit-code-v1-3b" && m != "wizard-mega-13b-awq" {
			return nil, ErrUnknownModel
		}

		llm := providers.NewReplicateProvider(ctx, m)
		return llm, nil
	default:
		return nil, ErrUnknownModel
	}
}
