package db

import (
	"encoding/json"
)

type Memory struct {
	ID     string `json:"id"`
	UserId string `json:"user_id"`
	Public bool   `json:"public"`
}
type MatchParams struct {
	QueryEmbedding []float64 `json:"query_embedding"`
	MatchTreshold  float64   `json:"match_threshold"`
	MatchCount     int16     `json:"match_count"`
	MemoryID       []string  `json:"memoryid"`
	UserID         string    `json:"userid"`
}

type MatchResult struct {
	ID         string  `json:"id"`
	Content    string  `json:"content"`
	Similarity float64 `json:"similarity"`
}

type Embedding struct {
	MemoryId  string    `json:"memory_id"`
	UserId    string    `json:"user_id"`
	Content   string    `json:"content"`
	Embedding []float64 `json:"embedding"`
}

func CreateMemory(memoryId string, userId string, public bool) error {
	client, err := CreateClient()
	if err != nil {
		return err
	}

	_, _, err = client.From("memories").
		Insert(Memory{ID: memoryId, UserId: userId, Public: public}, false, "", "", "exact").
		Execute()

	return err
}

func AddMemory(userId string, memoryId string, content string, embedding []float64) error {
	client, err := CreateClient()
	if err != nil {
		return err
	}

	_, _, err = client.From("embeddings").Insert(Embedding{
		UserId:    userId,
		MemoryId:  memoryId,
		Content:   content,
		Embedding: embedding,
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

func MatchEmbeddings(memoryIds []string, userId string, embedding []float64) ([]MatchResult, error) {
	params := MatchParams{
		QueryEmbedding: embedding,
		MatchTreshold:  0.70,
		MatchCount:     10,
		MemoryID:       memoryIds,
		UserID:         userId,
	}

	client, err := CreateClient()
	if err != nil {
		return nil, err
	}

	response := client.Rpc("retrieve_embeddings", "", params)

	var results []MatchResult
	err = json.Unmarshal([]byte(response), &results)

	if err != nil {
		return nil, err
	}

	if client.ClientError != nil {
		return nil, err
	}

	return results, nil
}
