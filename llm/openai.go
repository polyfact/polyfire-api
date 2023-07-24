package llm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
)

type TokenUsage struct {
	Input  int `json:"input"`
	Output int `json:"output"`
}

type Result struct {
	Result     string     `json:"result"`
	TokenUsage TokenUsage `json:"token_usage"`
}

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

	tokenUsage := CountTokens(content, model)

	if c != nil {
		(*c)(os.Getenv("OPENAI_MODEL"), tokenUsage)
	}

	return embeddings, nil
}

func Generate(task string, c *func(string, int, int)) (Result, error) {
	tokenUsage := TokenUsage{Input: 0, Output: 0}

	for i := 0; i < 5; i++ {
		log.Printf("Trying generation %d/5\n", i+1)
		llm, err := openai.NewChat()

		if err != nil {
			return Result{Result: "{\"error\":\"llm_init_failed\"}", TokenUsage: tokenUsage}, err
		}

		input_prompt := task
		ctx := context.Background()
		completion, err := llm.Call(ctx, []schema.ChatMessage{
			schema.HumanChatMessage{Text: input_prompt},
		})

		if err != nil {
			fmt.Printf("%v\n", err)
			continue
		}

		model := os.Getenv("OPENAI_MODEL")
		tokenUsage.Input += CountTokens(input_prompt, model)
		tokenUsage.Output += CountTokens(completion, model)

		if c != nil {
			(*c)(os.Getenv("OPENAI_MODEL"), tokenUsage.Input, tokenUsage.Output)
		}

		return Result{Result: completion, TokenUsage: tokenUsage}, err
	}

	return Result{Result: "{\"error\":\"generation_failed\"}", TokenUsage: tokenUsage}, errors.New("Generation failed after 5 retries")
}
