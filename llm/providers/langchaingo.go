package providers

import (
	"context"
	"errors"

	db "github.com/polyfire/api/db"
	"github.com/polyfire/api/llm/providers/options"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

type LangchainProvider struct {
	Model     interface{ GetNumTokens(string) int }
	ModelName string
}

func (m LangchainProvider) Call(prompt string, opts *options.ProviderOptions) (string, error) {
	ctx := context.Background()
	var result string
	var err error

	if opts == nil {
		opts = &options.ProviderOptions{}
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
	c options.ProviderCallback,
	opts *options.ProviderOptions,
) chan options.Result {
	chanRes := make(chan options.Result)

	go func(chanRes chan options.Result) {
		defer close(chanRes)
		tokenUsage := options.TokenUsage{Input: 0, Output: 0}
		input_prompt := task
		completion, err := m.Call(input_prompt, opts)
		if err != nil {
			chanRes <- options.Result{Err: "generation_error"}
			return
		}

		if c != nil {
			(*c)(
				"cohere",
				m.ModelName,
				m.Model.GetNumTokens(input_prompt),
				m.Model.GetNumTokens(completion),
				completion,
				nil,
			)
		}

		tokenUsage.Input += m.Model.GetNumTokens(input_prompt)
		tokenUsage.Output += m.Model.GetNumTokens(completion)

		result := options.Result{Result: completion, TokenUsage: tokenUsage}

		chanRes <- result
	}(chanRes)

	return chanRes
}

func (m LangchainProvider) UserAllowed(user_id string) bool {
	res, _ := db.ProjectIsPremium(user_id)
	return res
}

func (m LangchainProvider) Name() string {
	return m.ModelName
}

func (m LangchainProvider) ProviderModel() (string, string) {
	return m.ModelName, m.ModelName
}

func (m LangchainProvider) DoesFollowRateLimit() bool {
	return true
}
