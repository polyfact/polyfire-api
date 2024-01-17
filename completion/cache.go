package completion

import (
	"context"
	"log"

	database "github.com/polyfire/api/db"
	"github.com/polyfire/api/llm"
	"github.com/polyfire/api/llm/providers/options"
	"github.com/polyfire/api/utils"
)

func CheckFuzzyCache(
	ctx context.Context,
	prompt string,
	providerName string,
	modelName string,
) (chan options.Result, []float32, error) {
	db := ctx.Value(utils.ContextKeyDB).(database.Database)
	embeddings, err := llm.Embed(ctx, []string{prompt}, nil)
	if err != nil {
		return nil, embeddings[0], err
	}

	cache, err := db.GetCompletionCacheByInput(providerName, modelName, embeddings[0])
	if err != nil {
		return nil, embeddings[0], err
	}

	if cache != nil {
		log.Println("[INFO] Fuzzy cache hit")
		result := make(chan options.Result)
		go func() {
			defer close(result)
			result <- options.Result{Result: cache.Result}
		}()
		return result, embeddings[0], nil
	}

	return nil, embeddings[0], nil
}

func CheckExactCache(
	ctx context.Context,
	prompt string,
	providerName string,
	modelName string,
) (chan options.Result, error) {
	db := ctx.Value(utils.ContextKeyDB).(database.Database)
	cache, err := db.GetExactCompletionCacheByHash(providerName, modelName, prompt)
	if err != nil {
		return nil, err
	}

	if cache != nil {
		log.Println("[INFO] Exact cache hit")
		result := make(chan options.Result)
		go func() {
			defer close(result)
			result <- options.Result{Result: cache.Result}
		}()
		return result, nil
	}

	return nil, nil
}
