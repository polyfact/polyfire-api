package utils

import (
	db "github.com/polyfire/api/db"
	llmTokens "github.com/polyfire/api/tokens"
)

func CutChatHistory(chatMessages []db.ChatMessage, maxToken int) []db.ChatMessage {
	var res []db.ChatMessage
	tokens := 0

	for _, item := range chatMessages {
		textTokens := llmTokens.CountTokens(item.Content)

		if tokens+textTokens > maxToken {
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
	ContextKeyUserID                ContextKey = "userId"
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
	ContextKeyOriginDomain          ContextKey = "originDomain"
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
