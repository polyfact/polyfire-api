package memory

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	router "github.com/julienschmidt/httprouter"

	database "github.com/polyfire/api/db"
	"github.com/polyfire/api/llm"
	"github.com/polyfire/api/tokens"
	"github.com/polyfire/api/utils"
)

const BatchSize int = 512

func Create(w http.ResponseWriter, r *http.Request, _ router.Params) {
	db := r.Context().Value(utils.ContextKeyDB).(database.Database)
	record := r.Context().Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)
	userID, ok := r.Context().Value(utils.ContextKeyUserID).(string)
	if !ok {
		utils.RespondError(w, record, "user_id_missing")
		return
	}

	var requestBody struct {
		Public *bool `json:"public,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		utils.RespondError(w, record, "decode_error")
		return
	}

	if requestBody.Public == nil {
		defaultVal := false
		requestBody.Public = &defaultVal
	}

	memoryID := uuid.New().String()

	if err := db.CreateMemory(memoryID, userID, *requestBody.Public); err != nil {
		utils.RespondError(w, record, "db_creation_error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	memory := database.Memory{ID: memoryID, UserID: userID, Public: *requestBody.Public}

	response, _ := json.Marshal(&memory)
	record(string(response))

	if err := json.NewEncoder(w).Encode(memory); err != nil {
		utils.RespondError(w, record, "encode_error")
	}
}

func Add(w http.ResponseWriter, r *http.Request, _ router.Params) {
	db := r.Context().Value(utils.ContextKeyDB).(database.Database)
	decoder := json.NewDecoder(r.Body)
	record := r.Context().Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)
	userID := r.Context().Value(utils.ContextKeyUserID).(string)

	var requestBody struct {
		ID       string      `json:"id"`
		Input    interface{} `json:"input"`
		MaxToken int         `json:"max_token"`
	}

	err := decoder.Decode(&requestBody)
	if err != nil {
		utils.RespondError(w, record, "decode_error")
		return
	}

	var chunkSize int

	if requestBody.MaxToken > 0 {
		chunkSize = requestBody.MaxToken
	} else {
		chunkSize = BatchSize
	}

	chunks := make([]string, 0)

	inputs := utils.StringOptionalArray(requestBody.Input)

	if len(inputs) == 0 || (len(inputs) == 1 && len(inputs[0]) == 0) {
		utils.RespondError(w, record, "empty_input")
		return
	}

	for _, input := range inputs {
		chunk := tokens.SplitText(input, chunkSize)

		chunks = append(chunks, chunk[:]...)
	}

	callback := func(model_name string, input_count int) {
		db.LogRequests(
			r.Context().Value(utils.ContextKeyEventID).(string),
			userID, "openai", model_name, input_count, 0, "embedding", true)
	}

	var embeddings [][]float32

	batches, err := tokens.BatchText(chunks, 2000)
	if err != nil {
		utils.RespondError(w, record, "embedding_error")
		return
	}

	for _, batch := range batches {
		embeddingsBatch, err := llm.Embed(r.Context(), batch, &callback)
		if err != nil {
			utils.RespondError(w, record, "embedding_error")
			return
		}

		embeddings = append(embeddings, embeddingsBatch[:]...)
	}

	results := make([]database.Embedding, 0)

	for i, chunk := range chunks {
		results = append(results, database.Embedding{
			UserID:    userID,
			MemoryID:  requestBody.ID,
			Content:   chunk,
			Embedding: embeddings[i],
		})
	}

	err = db.AddMemories(requestBody.ID, results)

	if err != nil {
		utils.RespondError(w, record, "db_insert_error")
		return
	}

	response := map[string]bool{"success": true}

	w.Header().Set("Content-Type", "application/json")

	responseStr, _ := json.Marshal(&response)
	record(string(responseStr))

	_ = json.NewEncoder(w).Encode(response)
}

func Get(w http.ResponseWriter, r *http.Request, _ router.Params) {
	db := r.Context().Value(utils.ContextKeyDB).(database.Database)
	record := r.Context().Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)
	userID, ok := r.Context().Value(utils.ContextKeyUserID).(string)
	if !ok {
		utils.RespondError(w, record, "user_id_error")
		return
	}

	results, err := db.GetMemoryIDs(userID)
	if err != nil {
		utils.RespondError(w, record, "retrieval_error")
		return
	}

	ids := make([]string, len(results))
	for i, result := range results {
		ids[i] = result.ID
	}

	response := map[string][]string{"ids": ids}

	w.Header().Set("Content-Type", "application/json")

	responseStr, _ := json.Marshal(&response)
	record(string(responseStr))

	_ = json.NewEncoder(w).Encode(response)
}

func Embedder(ctx context.Context, userID string, memoryID []string, task string) ([]database.MatchResult, error) {
	db := ctx.Value(utils.ContextKeyDB).(database.Database)
	callback := func(model_name string, input_count int) {
		db.LogRequests(
			ctx.Value(utils.ContextKeyEventID).(string),
			userID, "openai", model_name, input_count, 0, "embedding", true)
	}

	embeddings, err := llm.Embed(ctx, []string{task}, &callback)
	if err != nil {
		return nil, err
	}

	results, err := db.MatchEmbeddings(memoryID, userID, embeddings[0])
	if err != nil {
		return nil, err
	}

	return results, nil
}

func Search(w http.ResponseWriter, r *http.Request, p router.Params) {
	id := p.ByName("id")
	decoder := json.NewDecoder(r.Body)
	record := r.Context().Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)
	userID := r.Context().Value(utils.ContextKeyUserID).(string)

	var requestBody struct {
		Input string `json:"input"`
	}

	err := decoder.Decode(&requestBody)
	if err != nil {
		utils.RespondError(w, record, "decode_error")
		return
	}

	results, err := Embedder(r.Context(), userID, []string{id}, requestBody.Input)
	if err != nil {
		utils.RespondError(w, record, "embedding_error")
		return
	}

	response := map[string][]database.MatchResult{"results": results}

	w.Header().Set("Content-Type", "application/json")

	responseStr, _ := json.Marshal(&response)
	record(string(responseStr))

	_ = json.NewEncoder(w).Encode(response)
}
