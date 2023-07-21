package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"

	supa "github.com/nedpals/supabase-go"
	embeddings "github.com/tmc/langchaingo/embeddings"
	textsplitter "github.com/tmc/langchaingo/textsplitter"

	rpcsupa "github.com/supabase/postgrest-go"
)

type Memory struct {
	ID      string `json:"id"`
	USER_ID string `json:"user_id"`
}

const BatchSize int = 512

func Create(w http.ResponseWriter, r *http.Request) {
	id := uuid.New().String()
	user_id := r.Context().Value("user_id").(string)

	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	client := supa.CreateClient(supabaseUrl, supabaseKey)

	var results []Memory

	err := client.DB.From("memories").Insert(Memory{ID: id, USER_ID: user_id}).Execute(&results)

	if err != nil {
		http.Error(w, fmt.Sprintf("500 Internal Server Error: %v", err), http.StatusInternalServerError)
		return
	}

	response := Memory{ID: id, USER_ID: user_id}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type Embedding struct {
	MEMORY_ID string    `json:"memory_id"`
	USER_ID   string    `json:"user_id"`
	CONTENT   string    `json:"content"`
	EMBEDDING []float64 `json:"embedding"`
}

func Add(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	user_id := r.Context().Value("user_id").(string)

	var requestBody struct {
		ID    string `json:"id"`
		Input string `json:"input"`
	}
	err := decoder.Decode(&requestBody)
	if err != nil {
		http.Error(w, "400 Bad Request", http.StatusBadRequest)
		return
	}

	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	client := supa.CreateClient(supabaseUrl, supabaseKey)

	value := os.Getenv("OPENAI_MODEL")

	os.Setenv("OPENAI_MODEL", "text-embedding-ada-002")

	embedder, err := embeddings.NewOpenAI(
		embeddings.WithStripNewLines(true),
		embeddings.WithBatchSize(BatchSize),
	)
	os.Setenv("OPENAI_MODEL", value)

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

	ctx := context.Background()
	for _, chunk := range chunks {
		embedding, err := embedder.EmbedQuery(ctx, chunk)
		if err != nil {
			http.Error(w, fmt.Sprintf("500 Internal Server Error: %v", err), http.StatusInternalServerError)
			return
		}

		err = client.DB.From("embeddings").Insert(Embedding{
			USER_ID:   user_id,
			MEMORY_ID: requestBody.ID,
			CONTENT:   chunk,
			EMBEDDING: embedding,
		}).Execute(nil)

		if err != nil {
			http.Error(w, fmt.Sprintf("500 Internal Server Error: %v", err), http.StatusInternalServerError)
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

var results []memoryRecord

func Get(w http.ResponseWriter, r *http.Request) {
	user_id, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusBadRequest)
		return
	}

	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	client := supa.CreateClient(supabaseUrl, supabaseKey)

	var results []memoryRecord
	err := client.DB.From("memories").Select("id").Eq("user_id", user_id).Execute(&results)

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

type Result struct {
	ID         string  `json:"id"`
	Content    string  `json:"content"`
	Similarity float64 `json:"similarity"`
}

func Embedder(task string) ([]Result, error) {
	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	fmt.Println(supabaseUrl, supabaseKey)
	client := rpcsupa.NewClient(supabaseUrl+"/rest/v1", "", nil)
	if client.ClientError != nil {
		panic(client.ClientError)
	}

	client.TokenAuth(supabaseKey)

	value := os.Getenv("OPENAI_MODEL")

	os.Setenv("OPENAI_MODEL", "text-embedding-ada-002")

	embedder, err := embeddings.NewOpenAI(
		embeddings.WithStripNewLines(true),
		embeddings.WithBatchSize(512),
	)
	os.Setenv("OPENAI_MODEL", value)

	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	embedding, err := embedder.EmbedQuery(ctx, task)
	if err != nil {
		return nil, err
	}

	params := map[string]interface{}{
		"query_embedding": embedding,
		"match_threshold": 0.70,
		"match_count":     10,
	}

	response := client.Rpc("match_embeddings", "", params)

	var results []Result
	err = json.Unmarshal([]byte(response), &results)

	if err != nil {
		log.Fatal(err)
	}

	if client.ClientError != nil {
		return nil, err
	}

	return results, nil
}
