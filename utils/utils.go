package utils

import (
	db "github.com/polyfire/api/db"
	llmTokens "github.com/polyfire/api/tokens"
)

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

func ContainsString(list []string, item string) bool {
	for _, i := range list {
		if i == item {
			return true
		}
	}

	return false
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
	ContextKeyProjectID             ContextKey = "projectID"
	ContextKeyElevenlabsToken       ContextKey = "elevenlabsToken"
	ContextKeyEventID               ContextKey = "eventID"
)

type EventType string

const (
	Unknown EventType = "unknown"

	AuthFirebase            EventType = "auth.token.firebase"
	AuthCustom              EventType = "auth.token.custom"
	AuthAnonymous           EventType = "auth.token.anonymous"
	AuthProviderRedirection EventType = "auth.provider.redirection"
	AuthProviderCallback    EventType = "auth.provider.callback"
	AuthProviderRefresh     EventType = "auth.provider.refresh"

	AuthID EventType = "auth.user.id"

	Usage EventType = "auth.user.usage"

	Generate    EventType = "models.completion.generate"
	ChatHistory EventType = "models.chat.history"
	ChatCreate  EventType = "models.chat.create"

	SpeechToText EventType = "models.stt.transcribe"
	TextToSpeech EventType = "models.tts.synthesize"

	ImageGeneration EventType = "models.image.generate"

	MemoryList   EventType = "data.memory.list"
	MemoryCreate EventType = "data.memory.create"
	MemoryAdd    EventType = "data.memory.add"
	MemorySearch EventType = "data.memory.search"

	KVGet    EventType = "data.kv.get"
	KVSet    EventType = "data.kv.set"
	KVDelete EventType = "data.kv.delete"
	KVList   EventType = "data.kv.list"

	PromptLike   EventType = "data.prompt.like"
	PromptGet    EventType = "data.prompt.get"
	PromptList   EventType = "data.prompt.list"
	PromptCreate EventType = "data.prompt.create"
	PromptUpdate EventType = "data.prompt.update"
	PromptDelete EventType = "data.prompt.delete"
)
