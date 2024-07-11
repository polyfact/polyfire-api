package db

import (
	"fmt"
	"os"

	postgrest "github.com/supabase/postgrest-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database interface {
	getUserInfos(userID string) (*UserInfos, error)
	CheckDBVersionRateLimit(
		userID string,
		version int,
	) (*UserInfos, RateLimitStatus, CreditsStatus, error)
	RemoveCreditsFromDev(userID string, credits int) error
	CreateRefreshToken(refreshToken string, refreshTokenSupabase string, projectID string) error
	GetAndDeleteRefreshToken(refreshToken string) (*RefreshToken, error)
	GetDevEmail(projectID string) (string, error)
	GetUserIDFromProjectAuthID(project string, authID string) (*string, error)
	CreateProjectUser(authID string, projectID string, monthlyCreditRateLimit *int) (*string, error)
	GetTTSVoice(slug string) (TTSVoice, error)
	GetCompletionCache(id string) (*CompletionCache, error)
	GetCompletionCacheByInput(
		provider string,
		model string,
		input []float32,
	) (*CompletionCache, error)
	AddCompletionCache(
		input []float32,
		prompt string,
		result string,
		provider string,
		model string,
		exact bool,
	) error
	GetExactCompletionCacheByHash(
		provider string,
		model string,
		input string,
	) (*CompletionCache, error)
	LogRequests(
		eventID string,
		userID string,
		providerName string,
		modelName string,
		inputTokenCount int,
		outputTokenCount int,
		kind Kind,
		countCredits bool,
	)
	LogRequestsCredits(
		eventID string,
		userID string,
		modelName string,
		credits int,
		inputTokenCount int,
		outputTokenCount int,
		kind Kind,
	)
	LogEvents(
		id string,
		path string,
		userID string,
		projectID string,
		requestBody string,
		responseBody string,
		error bool,
		promptID string,
		eventType string,
		orginDomain string,
	)
	SetKV(userID, key, value string) error
	GetKV(userID, key string) (*KVStore, error)
	GetKVMap(userID string, keys []string) (map[string]string, error)
	DeleteKV(userID, key string) error
	ListKV(userID string) ([]KVStore, error)
	GetPromptByIDOrSlug(id string) (*Prompt, error)
	RetrieveSystemPromptID(systemPromptIDOrSlug *string) (*string, error)
	GetChatByID(id string) (*Chat, error)
	CreateChat(
		userID string,
		systemPrompt *string,
		SystemPromptID *string,
		name *string,
	) (*Chat, error)
	ListChats(userID string) ([]ChatWithLatestMessage, error)
	DeleteChat(userID string, id string) error
	UpdateChat(userID string, id string, name string) (*Chat, error)
	GetChatMessages(
		userID string,
		chatID string,
		orderByDESC bool,
		limit int,
		offset int,
	) ([]ChatMessage, error)
	AddChatMessage(chatID string, isUserMessage bool, content string) error
	CreateMemory(memoryID string, userID string, public bool) error
	GetMemory(memoryID string) (*Memory, error)
	AddMemory(userID string, memoryID string, content string, embedding []float32) error
	AddMemories(memoryID string, embeddings []Embedding) error
	GetExistingEmbeddingFromContent(content string) (*[]float32, error)
	GetMemoryIDs(userID string) ([]MemoryRecord, error)
	MatchEmbeddings(memoryIDs []string, userID string, embedding []float32) ([]MatchResult, error)
	GetProjectByID(id string) (*Project, error)
	GetProjectUserByID(id string) (*ProjectUser, error)
	GetProjectForUserID(userID string) (*string, error)
	GetModelByAliasAndProjectID(alias string, projectID string, modelType string) (*Model, error)
}

type DB struct {
	sql gorm.DB
}

func InitDB() DB {
	db, err := gorm.Open(postgres.Open(os.Getenv("POSTGRES_URI")), &gorm.Config{})
	if err != nil {
		fmt.Println("POSTGRES_URI: ", os.Getenv("POSTGRES_URI"))
		panic("POSTGRES_URI: " + os.Getenv("POSTGRES_URI"))
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic("Failed to get SQL connection")
	}

	sqlDB.SetMaxOpenConns(50)

	return DB{sql: *db}
}

func CreateClient() (*postgrest.Client, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	client := postgrest.NewClient(supabaseURL+"/rest/v1", "", nil)

	if client.ClientError != nil {
		return nil, client.ClientError
	}

	client.TokenAuth(supabaseKey)

	return client, nil
}
