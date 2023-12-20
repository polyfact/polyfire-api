package completion

import (
	"context"
	"log"

	db "github.com/polyfire/api/db"
	llm "github.com/polyfire/api/llm"
	options "github.com/polyfire/api/llm/providers/options"
)

func CheckFuzzyCache(
	ctx context.Context,
	prompt string,
	providerName string,
	modelName string,
) (chan options.Result, []float32, error) {
	embeddings, err := llm.Embed(ctx, prompt, nil)
	if err != nil {
		return nil, embeddings, err
	}

	cache, err := db.GetCompletionCacheByInput(providerName, modelName, embeddings)
	if err != nil {
		return nil, embeddings, err
	}

	if cache != nil {
		log.Println("Fuzzy Cache hit")

		result := make(chan options.Result)
		go func() {
			defer close(result)
			result <- options.Result{Result: cache.Result}
		}()
		return result, embeddings, nil
	}

	return nil, embeddings, nil
}

func CheckExactCache(
	prompt string,
	providerName string,
	modelName string,
) (chan options.Result, error) {
	cache, err := db.GetExactCompletionCacheByHash(providerName, modelName, prompt)
	if err != nil {
		return nil, err
	}

	if cache != nil {
		log.Println("Cache hit")

		result := make(chan options.Result)
		go func() {
			defer close(result)
			result <- options.Result{Result: cache.Result}
		}()
		return result, nil
	}

	return nil, nil
}
