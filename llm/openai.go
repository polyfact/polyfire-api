package llm

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"

	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"

	type_parser "github.com/polyfact/api/type_parser"
)

func generateTypedPrompt(type_format string, task string) string {
	return "Your goal is to write a JSON object that will accomplish a specific task.\nThe string inside the JSON must be plain text, and not contain any markdown or HTML unless explicitely mentionned in the task.\nThe JSON object should follow this type:\n```\n" + type_format + "\n``` The task you must accomplish:\n" + task + "\n\nPlease only provide the JSON in a single json markdown code block with the keys described above. Do not include any other text.\nPlease make sure the JSON is a single line and does not contain any newlines outside of the strings."
}

func removeMarkdownCodeBlock(s string) string {
	return strings.TrimPrefix(strings.Trim(strings.TrimSpace(s), "`"), "json")
}

func Generate(type_format any, task string) (any, error) {
	type_string, err := type_parser.TypeToString(type_format, 0)
	if err != nil {
		return "", err
	}

	for i := 0; i < 5; i++ {
		log.Printf("Trying generation %d/5\n", i+1)
		llm, err := openai.NewChat()
		if err != nil {
			return "", err
		}
		ctx := context.Background()
		completion, err := llm.Call(ctx, []schema.ChatMessage{
			schema.HumanChatMessage{Text: generateTypedPrompt(type_string, task)},
		})
		if err != nil {
			return "", err
		}

		result_json := removeMarkdownCodeBlock(completion)

		if !json.Valid([]byte(result_json)) {
			log.Printf("FAILED ATTEMPT: %s\n", result_json)
			continue
		}
		log.Printf("SUCCESS ATTEMPT: %s\n", result_json)

		var result interface{}
		json.Unmarshal([]byte(result_json), &result)

		type_check := type_parser.CheckAgainstType(type_format, result)

		if !type_check {
			continue
		}

		return result, nil
	}

	return "", errors.New("Generation failed after 5 retries")
}
