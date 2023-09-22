package llm

import (
	"context"
	"log"

	llmTokens "github.com/polyfact/api/tokens"
	"github.com/tmc/langchaingo/llms/openai"
)

func Embed(content string, c *func(string, int)) ([][]float64, error) {
	model := "text-embedding-ada-002"

	llm, err := openai.New(openai.WithModel(model))
	if err != nil {
		log.Fatalf("failed to create LLM: %v", err)
	}

	ctx := context.Background()
	embeddings, err := llm.CreateEmbedding(ctx, []string{content})
	if err != nil {
		return nil, err
	}

	token_usage := llmTokens.CountTokens(model, content)

	if c != nil {
		(*c)(model, token_usage)
	}

	return embeddings, nil
}
