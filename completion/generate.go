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
	Model          string      `json:"model,omitempty"`
	MemoryID       interface{} `json:"memory_id,omitempty"`
	ChatID         *string     `json:"chat_id,omitempty"`
	Stop           *[]string   `json:"stop,omitempty"`
	Temperature    *float32    `json:"temperature,omitempty"`
	Stream         bool        `json:"stream,omitempty"`
	SystemPromptID *string     `json:"system_prompt_id,omitempty"`
	SystemPrompt   *string     `json:"system_prompt,omitempty"`
	WebRequest     bool        `json:"web,omitempty"`
	Language       *string     `json:"language,omitempty"`
	FuzzyCache     bool        `json:"fuzzy_cache,omitempty"`
	Cache          *bool       `json:"cache,omitempty"`
	Infos          bool        `json:"infos,omitempty"`
	AutoComplete   bool        `json:"auto_complete,omitempty"`
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
	provider, err := llm.NewProvider(ctx, input.Model)
	if errors.Is(err, llm.ErrUnknownModel) {
		return nil, ErrUnknownModelProvider
	}

	if err != nil {
		return nil, ErrInternalServerError
	}

	providerName, modelName := provider.ProviderModel()
	fmt.Println("Provider: " + providerName)

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

	/*
		If the autocomplete flag is on, we skip the question/answer prompt and put
		the LLM "cursor" at the end of the task, effectively asking it to complete
		the text instead of answering a question.

		This might not be enough for some models retrained to answer chat questions
		instead of just completing a text. The systemPrompt should also be ajusted.
	*/
	var prompt string
	if input.AutoComplete {
		prompt = getLanguageCompletion(input.Language) + contextString + "\n" + input.Task
	} else {
		prompt = getLanguageCompletion(input.Language) + contextString + "\nUser:\n" + input.Task + "\nYou:\n"
	}

	fmt.Println("Prompt: " + prompt)

	var result chan options.Result

	var embeddings []float32

	if input.Temperature != nil && *(input.Temperature) == 0.0 && (input.Cache == nil || *(input.Cache)) {
		result, err = CheckExactCache(prompt, providerName, modelName)
	}

	if err != nil {
		return nil, err
	}

	if result != nil {
		return &result, nil
	}

	// The fuzzy cache check for "close enough" embeddings.
	// It can reduce costs a lot in some cases but might lead to data leakage.
	// It should never be used in places with user personnal informations.
	if input.FuzzyCache {
		result, embeddings, err = CheckFuzzyCache(ctx, prompt, providerName, modelName)
	}

	if err != nil {
		return nil, err
	}

	if result != nil {
		return &result, nil
	}

	log.Println("Generate")
	resChan := provider.Generate(prompt, &callback, &opts)

	if input.AutoComplete {
		resChan = AddSpaceIfNeeded(prompt, resChan)
	}

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
		if (input.Temperature != nil && *(input.Temperature) == 0.0 && (input.Cache == nil || *(input.Cache))) ||
			input.FuzzyCache {
			_ = db.AddCompletionCache(embeddings, prompt, totalCompletion, providerName, modelName, input.FuzzyCache)
		}
	}()
	return &result, nil
}
