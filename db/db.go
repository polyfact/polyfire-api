package db

import (
	"encoding/json"
	"log"
	"os"

	supa "github.com/nedpals/supabase-go"
	rpc "github.com/supabase/postgrest-go"
)

type Kind string

const (
	Completion Kind = "completion"
	Embed      Kind = "embedding"
)

func CreateRpcClient() (*rpc.Client, error) {
	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	client := rpc.NewClient(supabaseUrl+"/rest/v1", "", nil)

	if client.ClientError != nil {
		return nil, client.ClientError
	}

	client.TokenAuth(supabaseKey)

	return client, nil
}

func CreateClient() *supa.Client {
	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	client := supa.CreateClient(supabaseUrl, supabaseKey)

	return client
}

type RequestLog struct {
	UserID           string `json:"user_id"`
	ModelName        string `json:"model_name"`
	InputTokenCount  *int   `json:"input_token_count"`
	OutputTokenCount *int   `json:"output_token_count"`
	Kind             Kind   `json:"kind"`
}

type Memory struct {
	ID      string `json:"id"`
	USER_ID string `json:"user_id"`
}
type MatchParams struct {
	QueryEmbedding []float64 `json:"query_embedding"`
	MatchTreshold  float64   `json:"match_threshold"`
	MatchCount     int16     `json:"match_count"`
	MemoryID       string    `json:"memory_id"`
}

type MatchResult struct {
	ID         string  `json:"id"`
	Content    string  `json:"content"`
	Similarity float64 `json:"similarity"`
}

type Embedding struct {
	MEMORY_ID string    `json:"memory_id"`
	USER_ID   string    `json:"user_id"`
	CONTENT   string    `json:"content"`
	EMBEDDING []float64 `json:"embedding"`
}

func toRef[T interface{}](n T) *T {
	return &n
}

func LogRequests(
	user_id string,
	model_name string,
	input_token_count int,
	output_token_count int,
	kind Kind,
) {
	supabase := CreateClient()

	if kind == "" {
		kind = "completion"
	}

	row := RequestLog{
		UserID:           user_id,
		ModelName:        model_name,
		Kind:             kind,
		InputTokenCount:  toRef(input_token_count),
		OutputTokenCount: toRef(output_token_count),
	}

	var results []RequestLog
	err := supabase.DB.From("request_logs").Insert(row).Execute(&results)

	if err != nil {
		panic(err)
	}
}

func CreateMemory(memoryId string, userId string) error {
	client := CreateClient()

	err := client.DB.From("memories").Insert(Memory{ID: memoryId, USER_ID: userId}).Execute(nil)

	return err
}

func AddMemory(userId string, memoryId string, content string, embedding []float64) error {
	client := CreateClient()

	err := client.DB.From("embeddings").Insert(Embedding{
		USER_ID:   userId,
		MEMORY_ID: memoryId,
		CONTENT:   content,
		EMBEDDING: embedding,
	}).Execute(nil)

	return err
}

type MemoryRecord struct {
	ID string `json:"id"`
}

func GetMemoryIds(userId string) ([]MemoryRecord, error) {
	client := CreateClient()

	var results []MemoryRecord

	err := client.DB.From("memories").Select("id").Eq("user_id", userId).Execute(&results)

	if err != nil {
		return nil, err
	}

	return results, nil
}

func MatchEmbeddings(memoryId string, embedding []float64) ([]MatchResult, error) {
	params := MatchParams{
		QueryEmbedding: embedding,
		MatchTreshold:  0.70,
		MatchCount:     10,
		MemoryID:       memoryId,
	}

	client, err := CreateRpcClient()

	response := client.Rpc("match_embeddings", "", params)

	var results []MatchResult
	err = json.Unmarshal([]byte(response), &results)

	if err != nil {
		log.Fatal(err)
	}

	if client.ClientError != nil {
		return nil, err
	}

	return results, nil
}
