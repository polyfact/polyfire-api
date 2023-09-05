package memory

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	router "github.com/julienschmidt/httprouter"
	textsplitter "github.com/tmc/langchaingo/textsplitter"

	db "github.com/polyfact/api/db"
	"github.com/polyfact/api/llm"
	"github.com/polyfact/api/utils"
)

const BatchSize int = 512

func Create(w http.ResponseWriter, r *http.Request, _ router.Params) {
	record := r.Context().Value("recordEvent").(utils.RecordFunc)
	userId, ok := r.Context().Value("user_id").(string)
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
		defaultVal := true
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
	record := r.Context().Value("recordEvent").(utils.RecordFunc)
	userId := r.Context().Value("user_id").(string)

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
		db.LogRequests(userId, "openai", model_name, input_count, 0, "embedding")
	}

	for _, chunk := range chunks {
		embedding, err := llm.Embed(chunk, &callback)
		if err != nil {
			utils.RespondError(w, record, "embedding_error")
			return
		}

		err = db.AddMemory(userId, requestBody.ID, chunk, embedding[0])

		if err != nil {
			utils.RespondError(w, record, "db_insert_error")
			return
		}
	}

	response := map[string]bool{"success": true}

	w.Header().Set("Content-Type", "application/json")

	response_str, _ := json.Marshal(&response)
	record(string(response_str))

	json.NewEncoder(w).Encode(response)
}

type memoryRecord struct {
	ID string `json:"id"`
}

func Get(w http.ResponseWriter, r *http.Request, _ router.Params) {
	record := r.Context().Value("recordEvent").(utils.RecordFunc)
	userId, ok := r.Context().Value("user_id").(string)
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

	json.NewEncoder(w).Encode(response)
}

func Embedder(userId string, memoryId string, task string) ([]db.MatchResult, error) {
	callback := func(model_name string, input_count int) {
		db.LogRequests(userId, "openai", model_name, input_count, 0, "embedding")
	}

	embeddings, err := llm.Embed(task, &callback)
	if err != nil {
		return nil, err
	}

	results, err := db.MatchEmbeddings(memoryId, userId, embeddings[0])
	if err != nil {
		return nil, err
	}

	return results, nil
}
