package completion

import (
	"context"

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
	embeddings, err := llm.Embed(ctx, prompt, nil)
	if err != nil {
		return nil, embeddings, err
	}

	cache, err := db.GetCompletionCacheByInput(providerName, modelName, embeddings)
	if err != nil {
		return nil, embeddings, err
	}

	if cache != nil {
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
		result := make(chan options.Result)
		go func() {
			defer close(result)
			result <- options.Result{Result: cache.Result}
		}()
		return result, nil
	}

	return nil, nil
}
