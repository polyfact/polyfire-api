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
	MockGetExistingEmbeddingFromContent func(content string) (*[]float32, error)
	MockGetMemoryIDs                    func(userID string) ([]MemoryRecord, error)
	MockMatchEmbeddings                 func(memoryIDs []string, userID string, embedding []float32) ([]MatchResult, error)
	MockGetProjectByID                  func(id string) (*Project, error)
	MockGetProjectUserByID              func(id string) (*ProjectUser, error)
	MockGetProjectForUserID             func(userID string) (*string, error)
	MockGetModelByAliasAndProjectID     func(alias string, projectID string, modelType string) (*Model, error)
}

func (mdb MockDatabase) GetModelByAliasAndProjectID(alias string, projectID string, modelType string) (*Model, error) {
	panic("Mock GetModelByAliasAndProjectID Unimplemented")
}

func (mdb MockDatabase) GetProjectForUserID(userID string) (*string, error) {
	panic("Mock GetProjectForUserID Unimplemented")
}

func (mdb MockDatabase) GetProjectUserByID(id string) (*ProjectUser, error) {
	panic("Mock GetProjectUserByID Unimplemented")
}

func (mdb MockDatabase) GetProjectByID(id string) (*Project, error) {
	panic("Mock GetProjectByID Unimplemented")
}

func (mdb MockDatabase) MatchEmbeddings(memoryIDs []string, userID string, embedding []float32) ([]MatchResult, error) {
	panic("Mock MatchEmbeddings Unimplemented")
}

func (mdb MockDatabase) GetMemoryIDs(userID string) ([]MemoryRecord, error) {
	panic("Mock GetMemoryIDs Unimplemented")
}

func (mdb MockDatabase) GetExistingEmbeddingFromContent(content string) (*[]float32, error) {
	panic("Mock GetExistingEmbeddingFromContent Unimplemented")
}

func (mdb MockDatabase) AddMemory(userID string, memoryID string, content string, embedding []float32) error {
	panic("Mock AddMemory Unimplemented")
}

func (mdb MockDatabase) GetMemory(memoryID string) (*Memory, error) {
	panic("Mock GetMemory Unimplemented")
}

func (mdb MockDatabase) CreateMemory(memoryID string, userID string, public bool) error {
	panic("Mock CreateMemory Unimplemented")
}

func (mdb MockDatabase) AddChatMessage(chatID string, isUserMessage bool, content string) error {
	panic("Mock AddChatMessage Unimplemented")
}

func (mdb MockDatabase) GetChatMessages(
	userID string,
	chatID string,
	orderByDESC bool,
	limit int,
	offset int,
) ([]ChatMessage, error) {
	panic("Mock GetChatMessages Unimplemented")
}

func (mdb MockDatabase) UpdateChat(userID string, id string, name string) (*Chat, error) {
	panic("Mock UpdateChat Unimplemented")
}

func (mdb MockDatabase) DeleteChat(userID string, id string) error {
	panic("Mock DeleteChat Unimplemented")
}

func (mdb MockDatabase) ListChats(userID string) ([]ChatWithLatestMessage, error) {
	panic("Mock ListChats Unimplemented")
}

func (mdb MockDatabase) CreateChat(userID string, systemPrompt *string, SystemPromptID *string, name *string) (*Chat, error) {
	panic("Mock CreateChat Unimplemented")
}

func (mdb MockDatabase) GetChatByID(id string) (*Chat, error) {
	panic("Mock GetChatByID Unimplemented")
}

func (mdb MockDatabase) RetrieveSystemPromptID(systemPromptIDOrSlug *string) (*string, error) {
	panic("Mock RetrieveSystemPromptID Unimplemented")
}

func (mdb MockDatabase) GetPromptByIDOrSlug(id string) (*Prompt, error) {
	panic("Mock GetPromptByIDOrSlug Unimplemented")
}

func (mdb MockDatabase) ListKV(userID string) ([]KVStore, error) {
	panic("Mock ListKV Unimplemented")
}

func (mdb MockDatabase) DeleteKV(userID, key string) error {
	panic("Mock DeleteKV Unimplemented")
}

func (mdb MockDatabase) GetKVMap(userID string, keys []string) (map[string]string, error) {
	panic("Mock GetKVMap Unimplemented")
}

func (mdb MockDatabase) GetKV(userID, key string) (*KVStore, error) {
	panic("Mock GetKV Unimplemented")
}

func (mdb MockDatabase) SetKV(userID, key, value string) error {
	panic("Mock SetKV Unimplemented")
}

func (mdb MockDatabase) LogEvents(
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
) {
	panic("Mock LogEvents Unimplemented")
}

func (mdb MockDatabase) LogRequestsCredits(
	eventID string,
	userID string,
	modelName string,
	credits int,
	inputTokenCount int,
	outputTokenCount int,
	kind Kind,
) {
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
	panic("Mock LogRequests Unimplemented")
}

func (mdb MockDatabase) GetExactCompletionCacheByHash(provider string, model string, input string) (*CompletionCache, error) {
	if mdb.MockGetExactCompletionCacheByHash != nil {
		return mdb.MockGetExactCompletionCacheByHash(provider, model, input)
	}
	panic("GetExactCompletionCacheByHash Mock not found")
}

func (mdb MockDatabase) AddCompletionCache(
	input []float32,
	prompt string,
	result string,
	provider string,
	model string,
	exact bool,
) error {
	panic("Mock AddCompletionCache Unimplemented")
}

func (mdb MockDatabase) GetCompletionCacheByInput(provider string, model string, input []float32) (*CompletionCache, error) {
	if mdb.MockGetCompletionCacheByInput != nil {
		return mdb.MockGetCompletionCacheByInput(provider, model, input)
	}
	panic("GetCompletionCacheByInput Mock not found")
}

func (mdb MockDatabase) GetCompletionCache(id string) (*CompletionCache, error) {
	panic("Mock GetCompletionCache Unimplemented")
}

func (mdb MockDatabase) GetTTSVoice(slug string) (TTSVoice, error) {
	panic("Mock GetTTSVoice Unimplemented")
}

func (mdb MockDatabase) CreateProjectUser(authID string, projectID string, monthlyCreditRateLimit *int) (*string, error) {
	panic("Mock CreateProjectUser Unimplemented")
}

func (mdb MockDatabase) GetUserIDFromProjectAuthID(project string, authID string) (*string, error) {
	panic("Mock GetUserIDFromProjectAuthID Unimplemented")
}

func (mdb MockDatabase) GetDevEmail(projectID string) (string, error) {
	panic("Mock GetDevEmail Unimplemented")
}

func (mdb MockDatabase) GetAndDeleteRefreshToken(refreshToken string) (*RefreshToken, error) {
	panic("Mock GetAndDeleteRefreshToken Unimplemented")
}

func (mdb MockDatabase) CreateRefreshToken(refreshToken string, refreshTokenSupabase string, projectID string) error {
	panic("Mock CreateRefreshToken Unimplemented")
}

func (mdb MockDatabase) RemoveCreditsFromDev(userID string, credits int) error {
	panic("Mock RemoveCreditsFromDev Unimplemented")
}

func (mdb MockDatabase) CheckDBVersionRateLimit(
	userID string,
	version int,
) (*UserInfos, RateLimitStatus, CreditsStatus, error) {
	panic("Mock CheckDBVersionRateLimit Unimplemented")
}

func (mdb MockDatabase) getUserInfos(userID string) (*UserInfos, error) {
	panic("Mock getUserInfos Unimplemented")
}
