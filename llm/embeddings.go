package llm

import (
	"context"
	"fmt"

	providers "github.com/polyfire/api/llm/providers"
	llmTokens "github.com/polyfire/api/tokens"
	goOpenai "github.com/sashabaranov/go-openai"

	"github.com/polyfire/api/utils"
)

func Embed(ctx context.Context, contents []string, c *func(string, int)) ([][]float32, error) {
	userID := ctx.Value(utils.ContextKeyUserID).(string)

	client := providers.NewOpenAIStreamProvider(ctx, fmt.Sprint(goOpenai.AdaEmbeddingV2)).Client

	embeddingCtx := context.Background()
	res, err := client.CreateEmbeddings(embeddingCtx, goOpenai.EmbeddingRequestStrings{
		Input: contents,
		Model: goOpenai.AdaEmbeddingV2,
		User:  userID,
	})
	if err != nil {
		return nil, err
	}

	var embeddings [][]float32

	for _, embed := range res.Data {
		embeddings = append(embeddings, embed.Embedding)
	}

	model := "text-embedding-ada-002"

	tokenUsage := 0
	for _, content := range contents {
		tokenUsage += llmTokens.CountTokens(content)
	}

	if c != nil {
		(*c)(model, tokenUsage)
	}

	return embeddings, nil
}
