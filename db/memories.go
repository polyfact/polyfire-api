package db

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

type Memory struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	Public bool   `json:"public"`
}

type MatchParams struct {
	QueryEmbedding []float32 `json:"query_embedding"`
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

type FloatArray []float32

func (o *FloatArray) Scan(src any) error {
	res := make([]float32, 0)
	str, ok := src.(string)
	if !ok {
		return errors.New("src value cannot cast to string")
	}
	stringArray := strings.Split(strings.Trim(str, "[]{}"), ",")

	for _, v := range stringArray {
		f, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return err
		}
		res = append(res, float32(f))
	}

	*o = res

	return nil
}

func (FloatArray) GormDataType() string {
	return "float[]"
}

type Embedding struct {
	MemoryID  string     `json:"memory_id"`
	UserID    string     `json:"user_id"`
	Content   string     `json:"content"`
	Embedding FloatArray `json:"embedding"`
}

func CreateMemory(memoryID string, userID string, public bool) error {
	err := DB.Exec("INSERT INTO memories (id, user_id, public) VALUES (?, ?::uuid, ?)", memoryID, userID, public).Error
	if err != nil {
		return err
	}

	return err
}

func GetMemory(memoryID string) (*Memory, error) {
	var memory Memory
	err := DB.First(&memory, "id = ?", memoryID).Error
	if err != nil {
		return nil, err
	}

	return &memory, nil
}

func AddMemory(userID string, memoryID string, content string, embedding []float32) error {
	memory, err := GetMemory(memoryID)
	if err != nil {
		return err
	}

	if memory == nil || memory.UserID != userID {
		return errors.New("memory not found")
	}

	embeddingstr := ""
	for _, v := range embedding {
		embeddingstr += strconv.FormatFloat(float64(v), 'f', 6, 64) + ","
	}
	embeddingstr = strings.TrimRight(embeddingstr, ",")

	err = DB.Exec(
		"INSERT INTO embeddings (memory_id, user_id, content, embedding) VALUES (?, ?::uuid, ?, string_to_array(?, ',')::float[])",
		memoryID,
		userID,
		content,
		embeddingstr,
	).Error
	if err != nil {
		return err
	}

	return err
}

func GetExistingEmbeddingFromContent(content string) (*[]float32, error) {
	var embeddings []Embedding
	err := DB.Find(&embeddings, "content = ?", content).Error
	if err != nil {
		return nil, err
	}

	if len(embeddings) == 0 {
		return nil, nil
	}

	res := []float32(embeddings[0].Embedding)
	return &res, nil
}

type MemoryRecord struct {
	ID string `json:"id"`
}

func (MemoryRecord) TableName() string {
	return "memories"
}

func GetMemoryIDs(userID string) ([]MemoryRecord, error) {
	var results []MemoryRecord

	err := DB.Find(&results, "user_id = ?", userID).Error
	if err != nil {
		return nil, err
	}

	return results, nil
}

func MatchEmbeddings(memoryIDs []string, userID string, embedding []float32) ([]MatchResult, error) {
	params := MatchParams{
		QueryEmbedding: embedding,
		MatchTreshold:  0.70,
		MatchCount:     10,
		MemoryID:       memoryIDs,
		UserID:         userID,
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
