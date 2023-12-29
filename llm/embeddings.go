package llm

import (
	"context"

	providers "github.com/polyfire/api/llm/providers"
	llmTokens "github.com/polyfire/api/tokens"
	goOpenai "github.com/sashabaranov/go-openai"

	database "github.com/polyfire/api/db"
	"github.com/polyfire/api/utils"
)

func Embed(ctx context.Context, content string, c *func(string, int)) ([]float32, error) {
	db := ctx.Value(utils.ContextKeyDB).(database.Database)
	alreadyExistingEmbedding, err := db.GetExistingEmbeddingFromContent(content)
	if err != nil {
		return nil, err
	}

	if alreadyExistingEmbedding != nil {
		return *alreadyExistingEmbedding, nil
	}

	userID := ctx.Value(utils.ContextKeyUserID).(string)

	client := providers.NewOpenAIStreamProvider(ctx, string(goOpenai.AdaEmbeddingV2)).Client

	embeddingCtx := context.Background()
	res, err := client.CreateEmbeddings(embeddingCtx, goOpenai.EmbeddingRequestStrings{
		Input: []string{content},
		Model: goOpenai.AdaEmbeddingV2,
		User:  userID,
	})
	if err != nil {
		return nil, err
	}

	embeddings := res.Data[0].Embedding

	model := "text-embedding-ada-002"

	tokenUsage := llmTokens.CountTokens(content)

	if c != nil {
		(*c)(model, tokenUsage)
	}

	return embeddings, nil
}
