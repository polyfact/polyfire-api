package utils

import (
	"os"

	db "github.com/polyfact/api/db"
	llm "github.com/polyfact/api/llm"
)

func FillContext(embeddings []db.MatchResult) (string, error) {
	context := ""
	tokens := 0

	for _, item := range embeddings {
		textTokens := llm.CountTokens(item.Content, os.Getenv("OPENAI_MODEL"))

		if tokens+textTokens > 2000 {
			break
		}

		context += "\n" + item.Content
		tokens += textTokens
	}

	if len(context) > 0 {
		context = "Context : " + context + "\n\n"
	}

	return context, nil

}
