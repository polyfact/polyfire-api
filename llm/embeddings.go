package llm

import (
	"context"
	"os"

	llmTokens "github.com/polyfire/api/tokens"
	goOpenai "github.com/sashabaranov/go-openai"

	"github.com/polyfire/api/db"
	"github.com/polyfire/api/utils"
)

func Embed(ctx context.Context, content string, c *func(string, int)) ([]float32, error) {
	alreadyExistingEmbedding, err := db.GetExistingEmbeddingFromContent(content)
	if err != nil {
		return nil, err
	}

	if alreadyExistingEmbedding != nil {
		return *alreadyExistingEmbedding, nil
	}

	var config goOpenai.ClientConfig

	userId := ctx.Value(utils.ContextKeyUserID).(string)

	customToken, ok := ctx.Value(utils.ContextKeyOpenAIToken).(string)
	if ok {
		config = goOpenai.DefaultConfig(customToken)
		customOrg, ok := ctx.Value(utils.ContextKeyOpenAIOrg).(string)
		if ok {
			config.OrgID = customOrg
		}
	} else {
		config = goOpenai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
		config.OrgID = os.Getenv("OPENAI_ORGANIZATION")
	}

	client := goOpenai.NewClientWithConfig(config)

	embeddingCtx := context.Background()
	res, err := client.CreateEmbeddings(embeddingCtx, goOpenai.EmbeddingRequestStrings{
		Input: []string{content},
		Model: goOpenai.AdaEmbeddingV2,
		User:  userId,
	})
	if err != nil {
		return nil, err
	}

	embeddings := res.Data[0].Embedding

	model := "text-embedding-ada-002"

	token_usage := llmTokens.CountTokens(content)

	if c != nil {
		(*c)(model, token_usage)
	}

	return embeddings, nil
}
