package utils

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
	ContextKeyCreditsStatus         ContextKey = "creditsStatus"
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
	ContextKeyProjectUserUsage      ContextKey = "projectUserUsage"
	ContextKeyProjectUserRateLimit  ContextKey = "projectUserRateLimit"
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
