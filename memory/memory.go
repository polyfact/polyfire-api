package memory

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	router "github.com/julienschmidt/httprouter"
	textsplitter "github.com/tmc/langchaingo/textsplitter"

	db "github.com/polyfire/api/db"
	"github.com/polyfire/api/llm"
	"github.com/polyfire/api/utils"
)

const BatchSize int = 512

func Create(w http.ResponseWriter, r *http.Request, _ router.Params) {
	record := r.Context().Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)
	userId, ok := r.Context().Value(utils.ContextKeyUserID).(string)
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

	memoryId := uuid.New().String()

	if err := db.CreateMemory(memoryId, userId, *requestBody.Public); err != nil {
		utils.RespondError(w, record, "db_creation_error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	memory := db.Memory{ID: memoryId, UserId: userId, Public: *requestBody.Public}

	response, _ := json.Marshal(&memory)
	record(string(response))

	if err := json.NewEncoder(w).Encode(memory); err != nil {
		utils.RespondError(w, record, "encode_error")
	}
}

func Add(w http.ResponseWriter, r *http.Request, _ router.Params) {
	decoder := json.NewDecoder(r.Body)
	record := r.Context().Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)
	userId := r.Context().Value(utils.ContextKeyUserID).(string)

	var requestBody struct {
		ID       string `json:"id"`
		Input    string `json:"input"`
		MaxToken int    `json:"max_token"`
	}

	err := decoder.Decode(&requestBody)
	if err != nil {
		utils.RespondError(w, record, "decode_error")
		return
	}

	splitter := textsplitter.NewTokenSplitter()

	if requestBody.MaxToken != 0 {
		splitter.ChunkSize = requestBody.MaxToken
	} else {
		splitter.ChunkSize = BatchSize
	}

	chunks, err := splitter.SplitText(requestBody.Input)
	if err != nil {
		utils.RespondError(w, record, "splitting_error")
		return
	}

	callback := func(model_name string, input_count int) {
		db.LogRequests(userId, "openai", model_name, input_count, 0, "embedding", true)
	}

	for _, chunk := range chunks {
		embeddings, err := llm.Embed(r.Context(), chunk, &callback)
		if err != nil {
			utils.RespondError(w, record, "embedding_error")
			return
		}

		err = db.AddMemory(userId, requestBody.ID, chunk, embeddings)

		if err != nil {
			utils.RespondError(w, record, "db_insert_error")
			return
		}
	}

	response := map[string]bool{"success": true}

	w.Header().Set("Content-Type", "application/json")

	response_str, _ := json.Marshal(&response)
	record(string(response_str))

	_ = json.NewEncoder(w).Encode(response)
}

func Get(w http.ResponseWriter, r *http.Request, _ router.Params) {
	record := r.Context().Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)
	userId, ok := r.Context().Value(utils.ContextKeyUserID).(string)
	if !ok {
		utils.RespondError(w, record, "user_id_error")
		return
	}

	results, err := db.GetMemoryIds(userId)
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

	response_str, _ := json.Marshal(&response)
	record(string(response_str))

	_ = json.NewEncoder(w).Encode(response)
}

func Embedder(ctx context.Context, userId string, memoryId []string, task string) ([]db.MatchResult, error) {
	callback := func(model_name string, input_count int) {
		db.LogRequests(userId, "openai", model_name, input_count, 0, "embedding", true)
	}

	embeddings, err := llm.Embed(ctx, task, &callback)
	if err != nil {
		return nil, err
	}

	results, err := db.MatchEmbeddings(memoryId, userId, embeddings)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func Search(w http.ResponseWriter, r *http.Request, p router.Params) {
	id := p.ByName("id")
	decoder := json.NewDecoder(r.Body)
	record := r.Context().Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)
	userId := r.Context().Value(utils.ContextKeyUserID).(string)

	var requestBody struct {
		Input string `json:"input"`
	}

	err := decoder.Decode(&requestBody)
	if err != nil {
		utils.RespondError(w, record, "decode_error")
		return
	}

	results, err := Embedder(r.Context(), userId, []string{id}, requestBody.Input)
	if err != nil {
		utils.RespondError(w, record, "embedding_error")
		return
	}

	response := map[string][]db.MatchResult{"results": results}

	w.Header().Set("Content-Type", "application/json")

	response_str, _ := json.Marshal(&response)
	record(string(response_str))

	_ = json.NewEncoder(w).Encode(response)
}
