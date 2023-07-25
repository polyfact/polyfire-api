package db

import (
	"encoding/json"
	"log"
)

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

func CreateMemory(memoryId string, userId string) error {
	client, err := CreateClient()
	if err != nil {
		return err
	}

	_, _, err = client.From("memories").Insert(Memory{ID: memoryId, USER_ID: userId}, false, "", "", "exact").Execute()

	return err
}

func AddMemory(userId string, memoryId string, content string, embedding []float64) error {
	client, err := CreateClient()
	if err != nil {
		return err
	}

	_, _, err = client.From("embeddings").Insert(Embedding{
		USER_ID:   userId,
		MEMORY_ID: memoryId,
		CONTENT:   content,
		EMBEDDING: embedding,
	}, false, "", "", "exact").Execute()

	return err
}

type MemoryRecord struct {
	ID string `json:"id"`
}

func GetMemoryIds(userId string) ([]MemoryRecord, error) {
	client, err := CreateClient()
	if err != nil {
		return nil, err
	}

	var results []MemoryRecord

	_, err = client.From("memories").Select("id", "exact", false).Eq("user_id", userId).ExecuteTo(&results)
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

	client, err := CreateClient()

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
