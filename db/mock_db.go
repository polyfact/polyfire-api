package db

type MockDatabase struct {
	MockgetUserInfos                    func(userID string) (*UserInfos, error)
	MockCheckDBVersionRateLimit         func(userID string, version int) (*UserInfos, RateLimitStatus, CreditsStatus, error)
	MockRemoveCreditsFromDev            func(userID string, credits int) error
	MockCreateRefreshToken              func(refreshToken string, refreshTokenSupabase string, projectID string) error
	MockGetAndDeleteRefreshToken        func(refreshToken string) (*RefreshToken, error)
	MockGetDevEmail                     func(projectID string) (string, error)
	MockGetUserIDFromProjectAuthID      func(project string, authID string) (*string, error)
	MockCreateProjectUser               func(authID string, projectID string, monthlyCreditRateLimit *int) (*string, error)
	MockGetTTSVoice                     func(slug string) (TTSVoice, error)
	MockGetCompletionCache              func(id string) (*CompletionCache, error)
	MockGetCompletionCacheByInput       func(provider string, model string, input []float32) (*CompletionCache, error)
	MockAddCompletionCache              func(input []float32, prompt string, result string, provider string, model string, exact bool) error
	MockGetExactCompletionCacheByHash   func(provider string, model string, input string) (*CompletionCache, error)
	MockLogRequests                     func(eventID string, userID string, providerName string, modelName string, inputTokenCount int, outputTokenCount int, kind Kind, countCredits bool)
	MockLogRequestsCredits              func(eventID string, userID string, modelName string, credits int, inputTokenCount int, outputTokenCount int, kind Kind)
	MockLogEvents                       func(id string, path string, userID string, projectID string, requestBody string, responseBody string, error bool, promptID string, eventType string, orginDomain string)
	MockSetKV                           func(userID, key, value string) error
	MockGetKV                           func(userID, key string) (*KVStore, error)
	MockGetKVMap                        func(userID string, keys []string) (map[string]string, error)
	MockDeleteKV                        func(userID, key string) error
	MockListKV                          func(userID string) ([]KVStore, error)
	MockGetPromptByIDOrSlug             func(id string) (*Prompt, error)
	MockRetrieveSystemPromptID          func(systemPromptIDOrSlug *string) (*string, error)
	MockGetChatByID                     func(id string) (*Chat, error)
	MockCreateChat                      func(userID string, systemPrompt *string, SystemPromptID *string, name *string) (*Chat, error)
	MockListChats                       func(userID string) ([]ChatWithLatestMessage, error)
	MockDeleteChat                      func(userID string, id string) error
	MockUpdateChat                      func(userID string, id string, name string) (*Chat, error)
	MockGetChatMessages                 func(userID string, chatID string, orderByDESC bool, limit int, offset int) ([]ChatMessage, error)
	MockAddChatMessage                  func(chatID string, isUserMessage bool, content string) error
	MockCreateMemory                    func(memoryID string, userID string, public bool) error
	MockGetMemory                       func(memoryID string) (*Memory, error)
	MockAddMemory                       func(userID string, memoryID string, content string, embedding []float32) error
	MockAddMemories                     func(memoryID string, embeddings []Embedding) error
	MockGetExistingEmbeddingFromContent func(content string) (*[]float32, error)
	MockGetMemoryIDs                    func(userID string) ([]MemoryRecord, error)
	MockMatchEmbeddings                 func(memoryIDs []string, userID string, embedding []float32) ([]MatchResult, error)
	MockGetProjectByID                  func(id string) (*Project, error)
	MockGetProjectUserByID              func(id string) (*ProjectUser, error)
	MockGetProjectForUserID             func(userID string) (*string, error)
	MockGetModelByAliasAndProjectID     func(alias string, projectID string, modelType string) (*Model, error)
}

func (mdb MockDatabase) GetModelByAliasAndProjectID(_ string, _ string, _ string) (*Model, error) {
	panic("Mock GetModelByAliasAndProjectID Unimplemented")
}

func (mdb MockDatabase) GetProjectForUserID(_ string) (*string, error) {
	panic("Mock GetProjectForUserID Unimplemented")
}

func (mdb MockDatabase) GetProjectUserByID(_ string) (*ProjectUser, error) {
	panic("Mock GetProjectUserByID Unimplemented")
}

func (mdb MockDatabase) GetProjectByID(_ string) (*Project, error) {
	panic("Mock GetProjectByID Unimplemented")
}

func (mdb MockDatabase) MatchEmbeddings(memoryIDs []string, userID string, embedding []float32) ([]MatchResult, error) {
	if mdb.MockMatchEmbeddings != nil {
		return mdb.MockMatchEmbeddings(memoryIDs, userID, embedding)
	}
	panic("Mock MatchEmbeddings Unimplemented")
}

func (mdb MockDatabase) GetMemoryIDs(_ string) ([]MemoryRecord, error) {
	panic("Mock GetMemoryIDs Unimplemented")
}

func (mdb MockDatabase) GetExistingEmbeddingFromContent(content string) (*[]float32, error) {
	if mdb.MockGetExistingEmbeddingFromContent != nil {
		return mdb.MockGetExistingEmbeddingFromContent(content)
	}
	panic("Mock GetExistingEmbeddingFromContent Unimplemented")
}

func (mdb MockDatabase) AddMemory(_ string, _ string, _ string, _ []float32) error {
	panic("Mock AddMemory Unimplemented")
}

func (mdb MockDatabase) AddMemories(_ string, _ []Embedding) error {
	panic("Mock AddMemories Unimplemented")
}

func (mdb MockDatabase) GetMemory(_ string) (*Memory, error) {
	panic("Mock GetMemory Unimplemented")
}

func (mdb MockDatabase) CreateMemory(_ string, _ string, _ bool) error {
	panic("Mock CreateMemory Unimplemented")
}

func (mdb MockDatabase) AddChatMessage(_ string, _ bool, _ string) error {
	panic("Mock AddChatMessage Unimplemented")
}

func (mdb MockDatabase) GetChatMessages(_ string, _ string, _ bool, _ int, _ int) ([]ChatMessage, error) {
	panic("Mock GetChatMessages Unimplemented")
}

func (mdb MockDatabase) UpdateChat(_ string, _ string, _ string) (*Chat, error) {
	panic("Mock UpdateChat Unimplemented")
}

func (mdb MockDatabase) DeleteChat(_ string, _ string) error {
	panic("Mock DeleteChat Unimplemented")
}

func (mdb MockDatabase) ListChats(_ string) ([]ChatWithLatestMessage, error) {
	panic("Mock ListChats Unimplemented")
}

func (mdb MockDatabase) CreateChat(_ string, _ *string, _ *string, _ *string) (*Chat, error) {
	panic("Mock CreateChat Unimplemented")
}

func (mdb MockDatabase) GetChatByID(_ string) (*Chat, error) {
	panic("Mock GetChatByID Unimplemented")
}

func (mdb MockDatabase) RetrieveSystemPromptID(_ *string) (*string, error) {
	panic("Mock RetrieveSystemPromptID Unimplemented")
}

func (mdb MockDatabase) GetPromptByIDOrSlug(_ string) (*Prompt, error) {
	panic("Mock GetPromptByIDOrSlug Unimplemented")
}

func (mdb MockDatabase) ListKV(_ string) ([]KVStore, error) {
	panic("Mock ListKV Unimplemented")
}

func (mdb MockDatabase) DeleteKV(_ string, _ string) error {
	panic("Mock DeleteKV Unimplemented")
}

func (mdb MockDatabase) GetKVMap(_ string, _ []string) (map[string]string, error) {
	panic("Mock GetKVMap Unimplemented")
}

func (mdb MockDatabase) GetKV(_ string, _ string) (*KVStore, error) {
	panic("Mock GetKV Unimplemented")
}

func (mdb MockDatabase) SetKV(_ string, _ string, _ string) error {
	panic("Mock SetKV Unimplemented")
}

func (mdb MockDatabase) LogEvents(
	_ string,
	_ string,
	_ string,
	_ string,
	_ string,
	_ string,
	_ bool,
	_ string,
	_ string,
	_ string,
) {
	panic("Mock LogEvents Unimplemented")
}

func (mdb MockDatabase) LogRequestsCredits(_ string, _ string, _ string, _ int, _ int, _ int, _ Kind) {
	panic("Mock LogRequestsCredits Unimplemented")
}

func (mdb MockDatabase) LogRequests(
	eventID string,
	userID string,
	providerName string,
	modelName string,
	inputTokenCount int,
	outputTokenCount int,
	kind Kind,
	countCredits bool,
) {
	if mdb.MockLogRequests != nil {
		mdb.MockLogRequests(
			eventID,
			userID,
			providerName,
			modelName,
			inputTokenCount,
			outputTokenCount,
			kind,
			countCredits,
		)
		return
	}
	panic("Mock LogRequests Unimplemented")
}

func (mdb MockDatabase) GetExactCompletionCacheByHash(provider string, model string, input string) (*CompletionCache, error) {
	if mdb.MockGetExactCompletionCacheByHash != nil {
		return mdb.MockGetExactCompletionCacheByHash(provider, model, input)
	}
	panic("GetExactCompletionCacheByHash Mock not found")
}

func (mdb MockDatabase) AddCompletionCache(_ []float32, _ string, _ string, _ string, _ string, _ bool) error {
	panic("Mock AddCompletionCache Unimplemented")
}

func (mdb MockDatabase) GetCompletionCacheByInput(provider string, model string, input []float32) (*CompletionCache, error) {
	if mdb.MockGetCompletionCacheByInput != nil {
		return mdb.MockGetCompletionCacheByInput(provider, model, input)
	}
	panic("GetCompletionCacheByInput Mock not found")
}

func (mdb MockDatabase) GetCompletionCache(_ string) (*CompletionCache, error) {
	panic("Mock GetCompletionCache Unimplemented")
}

func (mdb MockDatabase) GetTTSVoice(_ string) (TTSVoice, error) {
	panic("Mock GetTTSVoice Unimplemented")
}

func (mdb MockDatabase) CreateProjectUser(_ string, _ string, _ *int) (*string, error) {
	panic("Mock CreateProjectUser Unimplemented")
}

func (mdb MockDatabase) GetUserIDFromProjectAuthID(project string, authID string) (*string, error) {
	if mdb.MockGetUserIDFromProjectAuthID != nil {
		return mdb.MockGetUserIDFromProjectAuthID(project, authID)
	}
	panic("Mock GetUserIDFromProjectAuthID Unimplemented")
}

func (mdb MockDatabase) GetDevEmail(_ string) (string, error) {
	panic("Mock GetDevEmail Unimplemented")
}

func (mdb MockDatabase) GetAndDeleteRefreshToken(_ string) (*RefreshToken, error) {
	panic("Mock GetAndDeleteRefreshToken Unimplemented")
}

func (mdb MockDatabase) CreateRefreshToken(_ string, _ string, _ string) error {
	panic("Mock CreateRefreshToken Unimplemented")
}

func (mdb MockDatabase) RemoveCreditsFromDev(_ string, _ int) error {
	panic("Mock RemoveCreditsFromDev Unimplemented")
}

func (mdb MockDatabase) CheckDBVersionRateLimit(_ string, _ int) (*UserInfos, RateLimitStatus, CreditsStatus, error) {
	panic("Mock CheckDBVersionRateLimit Unimplemented")
}

func (mdb MockDatabase) getUserInfos(_ string) (*UserInfos, error) {
	panic("Mock getUserInfos Unimplemented")
}
