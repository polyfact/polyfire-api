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

		if c != nil {
			(*c)(os.Getenv("OPENAI_MODEL"), llm.GetNumTokens(input_prompt), llm.GetNumTokens(completion))
		}

		tokenUsage.Input += llm.GetNumTokens(input_prompt)
		tokenUsage.Output += llm.GetNumTokens(completion)

		return Result{Result: completion, TokenUsage: tokenUsage}, err
	}

	return Result{Result: "{\"error\":\"generation_failed\"}", TokenUsage: tokenUsage}, errors.New("Generation failed after 5 retries")
}
