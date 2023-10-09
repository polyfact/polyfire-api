package utils

import (
	db "github.com/polyfire/api/db"
	llmTokens "github.com/polyfire/api/tokens"
)

func FillContext(embeddings []db.MatchResult) (string, error) {
	context := ""
	tokens := 0

	for _, item := range embeddings {
		textTokens := llmTokens.CountTokens("gpt3.5-turbo", item.Content)

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
		textTokens := llmTokens.CountTokens("gpt3.5-turbo", item.Content)

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
	ContextKeyRateLimitStatus       ContextKey = "rateLimitStatus"
	ContextKeyRecordEvent           ContextKey = "recordEvent"
	ContextKeyRecordEventWithUserID ContextKey = "recordEventWithUserID"
	ContextKeyRecordEventRequest    ContextKey = "recordEventRequest"
	ContextKeyOpenAIToken           ContextKey = "openAIToken"
	ContextKeyOpenAIOrg             ContextKey = "openAIOrg"
	ContextKeyReplicateToken        ContextKey = "replicateToken"
)
