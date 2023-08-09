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
	memoryId := uuid.New().String()
	userId := r.Context().Value("user_id").(string)

	err := db.CreateMemory(memoryId, userId)
	if err != nil {
		utils.RespondError(w, "db_creation_error")
		return
	}

	response := db.Memory{ID: memoryId, USER_ID: userId}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func Add(w http.ResponseWriter, r *http.Request, _ router.Params) {
	decoder := json.NewDecoder(r.Body)
	userId := r.Context().Value("user_id").(string)

	var requestBody struct {
		ID       string `json:"id"`
		Input    string `json:"input"`
		MaxToken int    `json:"max_token"`
	}

	err := decoder.Decode(&requestBody)
	if err != nil {
		utils.RespondError(w, "decode_error")
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
		utils.RespondError(w, "splitting_error")
		return
	}

	callback := func(model_name string, input_count int) {
		db.LogRequests(userId, model_name, input_count, 0, "embedding")
	}

	for _, chunk := range chunks {
		embedding, err := llm.Embed(chunk, &callback)
		if err != nil {
			utils.RespondError(w, "embedding_error")
			return
		}

		err = db.AddMemory(userId, requestBody.ID, chunk, embedding[0])

		if err != nil {
			utils.RespondError(w, "db_insert_error")
			return
		}
	}

	response := map[string]bool{"success": true}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type memoryRecord struct {
	ID string `json:"id"`
}

func Get(w http.ResponseWriter, r *http.Request, _ router.Params) {
	userId, ok := r.Context().Value("user_id").(string)
	if !ok {
		utils.RespondError(w, "user_id_error")
		return
	}

	results, err := db.GetMemoryIds(userId)
	if err != nil {
		utils.RespondError(w, "retrieval_error")
		return
	}

	ids := make([]string, len(results))
	for i, result := range results {
		ids[i] = result.ID
	}

	response := map[string][]string{"ids": ids}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func Embedder(userId string, memoryId string, task string) ([]db.MatchResult, error) {
	callback := func(model_name string, input_count int) {
		db.LogRequests(userId, model_name, input_count, 0, "embedding")
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
