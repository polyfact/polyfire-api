package providers

import (
	"context"
	"os"

	goOpenai "github.com/sashabaranov/go-openai"
)

/*
 * The Open router API is fully compatible with the OpenAI API interface, we can
 * just reuse it with a different baseURL
 */

func NewOpenRouterProvider(_ context.Context, model string) OpenAIStreamProvider {
	var config goOpenai.ClientConfig
	var isCustomToken bool

	config = goOpenai.DefaultConfig(os.Getenv("OPENROUTER_API_KEY"))
	isCustomToken = false
	config.BaseURL = "https://openrouter.ai/api/v1"

	return OpenAIStreamProvider{
		client:        *goOpenai.NewClientWithConfig(config),
		Model:         model,
		IsCustomToken: isCustomToken,
		Provider:      "openrouter",
	}
}
