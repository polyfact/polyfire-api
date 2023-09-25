package providers

import (
	"context"
	"errors"
	"fmt"
	"log"

	db "github.com/polyfact/api/db"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

type LangchainProvider struct {
	Model     interface{ GetNumTokens(string) int }
	ModelName string
}

func (m LangchainProvider) Call(prompt string, opts *ProviderOptions) (string, error) {
	ctx := context.Background()
	var result string
	var err error

	if opts == nil {
		opts = &ProviderOptions{}
	}

	options := llms.CallOptions{}

	if opts.StopWords != nil {
		options.StopWords = *opts.StopWords
	}

	if llm, ok := m.Model.(llms.LLM); ok {
		result, err = llm.Call(ctx, prompt, llms.WithOptions(options))
	} else if chat, ok := m.Model.(llms.ChatLLM); ok {
		var chatMessage *schema.AIChatMessage
		chatMessage, err = chat.Call(ctx, []schema.ChatMessage{
			schema.HumanChatMessage{Content: prompt},
		}, llms.WithOptions(options))
		result = chatMessage.Content
	} else {
		return "", errors.New("Model is neither LLM nor Chat")
	}

	if err != nil {
		return "", err
	}

	return result, nil
}

func (m LangchainProvider) Generate(
	task string,
	c ProviderCallback,
	opts *ProviderOptions,
) chan Result {
	chan_res := make(chan Result)

	go func(chan_res chan Result) {
		defer close(chan_res)
		tokenUsage := TokenUsage{Input: 0, Output: 0}
		for i := 0; i < 5; i++ {
			log.Printf("Trying generation %d/5\n", i+1)

			input_prompt := task
			completion, err := m.Call(input_prompt, opts)
			if err != nil {
				fmt.Printf("%v\n", err)
				continue
			}

			if c != nil {
				(*c)("cohere", m.ModelName, m.Model.GetNumTokens(input_prompt), m.Model.GetNumTokens(completion), completion)
			}

			tokenUsage.Input += m.Model.GetNumTokens(input_prompt)
			tokenUsage.Output += m.Model.GetNumTokens(completion)

			result := Result{Result: completion, TokenUsage: tokenUsage}

			chan_res <- result
			return
		}
		chan_res <- Result{
			Result:     "{\"error\":\"generation_failed\"}",
			TokenUsage: tokenUsage,
			Err: errors.New(
				"Generation failed after 5 retries",
			),
		}
	}(chan_res)

	return chan_res
}

func (m LangchainProvider) UserAllowed(user_id string) bool {
	res, _ := db.ProjectIsPremium(user_id)
	return res
}

func (m LangchainProvider) Name() string {
	return m.ModelName
}

func (m LangchainProvider) DoesFollowRateLimit() bool {
	return true
}
