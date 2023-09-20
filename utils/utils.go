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

const PROMPT_IDENTIFIER_TOKEN_COUNT int = 5

func CutChatHistory(chat_messages []db.ChatMessage, max_token int) []db.ChatMessage {
	var res []db.ChatMessage
	tokens := 0

	for _, item := range chat_messages {
		textTokens := llm.CountTokens(item.Content, os.Getenv("OPENAI_MODEL"))

		if tokens+textTokens > max_token {
			break
		}

		res = append(res, item)
		tokens += textTokens
	}

	return res
}

type ContextKey string

const (
	ContextKeyUserID                ContextKey = "user_id"
	ContextKeyRecordEvent           ContextKey = "recordEvent"
	ContextKeyRecordEventWithUserID ContextKey = "recordEventWithUserID"
	ContextKeyRecordEventRequest    ContextKey = "recordEventRequest"
)
