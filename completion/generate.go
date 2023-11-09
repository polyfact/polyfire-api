package completion

import (
	"context"
	"errors"
	"fmt"
	"log"

	db "github.com/polyfire/api/db"
	llm "github.com/polyfire/api/llm"
	options "github.com/polyfire/api/llm/providers/options"
	utils "github.com/polyfire/api/utils"
)

type GenerateRequestBody struct {
	Task           string      `json:"task"`
	Provider       string      `json:"provider,omitempty"`
	Model          *string     `json:"model,omitempty"`
	MemoryID       interface{} `json:"memory_id,omitempty"`
	ChatID         *string     `json:"chat_id,omitempty"`
	Stop           *[]string   `json:"stop,omitempty"`
	Temperature    *float32    `json:"temperature,omitempty"`
	Stream         bool        `json:"stream,omitempty"`
	SystemPromptID *string     `json:"system_prompt_id,omitempty"`
	SystemPrompt   *string     `json:"system_prompt,omitempty"`
	WebRequest     bool        `json:"web,omitempty"`
	Language       *string     `json:"language,omitempty"`
	Cache          bool        `json:"cache,omitempty"`
	Infos          bool        `json:"infos,omitempty"`
}

func getLanguageCompletion(language *string) string {
	if language != nil && *language != "" {
		return "Answer in " + *language + ".\n"
	}
	return ""
}

func GenerationStart(ctx context.Context, userID string, input GenerateRequestBody) (*chan options.Result, error) {
	resources := []db.MatchResult{}

	log.Println("Init provider")

	// Get provider
	provider, err := llm.NewProvider(ctx, input.Provider, input.Model)
	if errors.Is(err, llm.ErrUnknownModel) {
		return nil, ErrUnknownModelProvider
	}

	if err != nil {
		return nil, ErrInternalServerError
	}

	providerName, modelName := provider.ProviderModel()

	// Check Rate Limit
	if provider.DoesFollowRateLimit() {
		log.Println("Check Rate Limit")
		err = CheckRateLimit(ctx)
		if err != nil {
			return nil, err
		}
	}

	// Init log request callbacks
	callback := func(providerName string, modelName string, inputCount int, outputCount int, _ string, credit *int) {
		if credit != nil && provider.DoesFollowRateLimit() {
			db.LogRequestsCredits(
				ctx.Value(utils.ContextKeyEventID).(string),
				userID, modelName, *credit, inputCount, outputCount, "completion")
		} else {
			db.LogRequests(
				ctx.Value(utils.ContextKeyEventID).(string),
				userID,
				providerName,
				modelName,
				inputCount,
				outputCount,
				"completion",
				provider.DoesFollowRateLimit(),
			)
		}
	}

	// Get Options
	opts := options.ProviderOptions{}
	if input.Stop != nil {
		opts.StopWords = input.Stop
	}
	if input.Temperature != nil {
		opts.Temperature = input.Temperature
	}

	// Get Context elements
	contextString, warnings, err := GetContextString(ctx, userID, input, &callback, &opts)
	if err != nil {
		return nil, err
	}

	// Get Language prompt
	prompt := getLanguageCompletion(input.Language) + contextString + "\nUser:\n" + input.Task + "\nYou:\n"

	fmt.Println("Prompt: " + prompt)

	var result chan options.Result
	var embeddings []float32

	// Check for cache hits
	if input.Cache {
		result, embeddings, err = CheckCache(ctx, prompt, providerName, modelName)
		if err != nil {
			return nil, err
		}
		if result != nil {
			return &result, nil
		}
	}

	// Generate
	log.Println("Generate")
	resChan := provider.Generate(prompt, &callback, &opts)

	result = make(chan options.Result)

	// Add warnings and cache at the end of the generation
	go func() {
		defer close(result)
		totalCompletion := ""
		for res := range resChan {
			result <- res
			totalCompletion += res.Result
		}
		result <- options.Result{Resources: resources, Warnings: warnings}
		if input.Cache {
			_ = db.AddCompletionCache(embeddings, totalCompletion, providerName, modelName)
		}
	}()
	return &result, nil
}
