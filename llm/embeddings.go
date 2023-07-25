package llm

import (
	"context"
	"log"
	"os"

	"github.com/tmc/langchaingo/llms/openai"
)

func CountTokens(content string, model string) int {
	llm, err := openai.New(openai.WithModel(model))
	if err != nil {
		log.Println("Error initializing LLM for CountTokens :", err)
		return 0
	}
	return llm.GetNumTokens(content)
}

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

	token_usage := CountTokens(content, model)

	if c != nil {
		(*c)(os.Getenv("OPENAI_MODEL"), token_usage)
	}

	return embeddings, nil
}
