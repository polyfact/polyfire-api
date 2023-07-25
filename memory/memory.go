package memory

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	textsplitter "github.com/tmc/langchaingo/textsplitter"

	db "github.com/polyfact/api/db"
	"github.com/polyfact/api/llm"
)

const BatchSize int = 512

func Create(w http.ResponseWriter, r *http.Request) {
	memoryId := uuid.New().String()
	userId := r.Context().Value("user_id").(string)

	err := db.CreateMemory(memoryId, userId)

	if err != nil {
		http.Error(w, fmt.Sprintf("500 Internal Server Error: %v", err), http.StatusInternalServerError)
		return
	}

	response := db.Memory{ID: memoryId, USER_ID: userId}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func Add(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId := r.Context().Value("user_id").(string)

	var requestBody struct {
		ID    string `json:"id"`
		Input string `json:"input"`
	}
	err := decoder.Decode(&requestBody)
	if err != nil {
		http.Error(w, "400 Bad Request", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("500 Internal Server Error: %v", err), http.StatusInternalServerError)
		return
	}

	splitter := textsplitter.NewTokenSplitter()
	splitter.ChunkSize = BatchSize

	chunks, err := splitter.SplitText(requestBody.Input)
	if err != nil {
		http.Error(w, fmt.Sprintf("500 Internal Server Error: %v", err), http.StatusInternalServerError)
		return
	}

	callback := func(model_name string, input_count int) {
		db.LogRequests(userId, model_name, input_count, 0, "embedding")
	}

	for _, chunk := range chunks {
		embedding, err := llm.Embed(chunk, &callback)
		if err != nil {
			http.Error(w, fmt.Sprintf("500 Internal Server Error: %v", err), http.StatusInternalServerError)
			return
		}

		err = db.AddMemory(userId, requestBody.ID, chunk, embedding[0])

		if err != nil {
			http.Error(w, fmt.Sprintf("500 Internal Server Error: %v", err), http.StatusInternalServerError)
			return
		}

		fmt.Printf("%v\n", embedding)
	}

	response := map[string]bool{"success": true}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type memoryRecord struct {
	ID string `json:"id"`
}

func Get(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusBadRequest)
		return
	}

	results, err := db.GetMemoryIds(userId)

	if err != nil {
		http.Error(w, fmt.Sprintf("500 Internal Server Error: %v", err), http.StatusInternalServerError)
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

	results, err := db.MatchEmbeddings(memoryId, embeddings[0])

	if err != nil {
		return nil, err
	}

	return results, nil
}
