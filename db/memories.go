package db

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

type Memory struct {
	ID     string `json:"id"`
	UserId string `json:"user_id"`
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
	MemoryId  string     `json:"memory_id"`
	UserId    string     `json:"user_id"`
	Content   string     `json:"content"`
	Embedding FloatArray `json:"embedding"`
}

func CreateMemory(memoryId string, userId string, public bool) error {
	err := DB.Exec("INSERT INTO memories (id, user_id, public) VALUES (?, ?::uuid, ?)", memoryId, userId, public).Error
	if err != nil {
		return err
	}

	return err
}

func GetMemory(memoryId string) (*Memory, error) {
	var memory Memory
	err := DB.First(&memory, "id = ?", memoryId).Error
	if err != nil {
		return nil, err
	}

	return &memory, nil
}

func AddMemory(userId string, memoryId string, content string, embedding []float32) error {
	memory, err := GetMemory(memoryId)
	if err != nil {
		return err
	}

	if memory == nil || memory.UserId != userId {
		return errors.New("memory not found")
	}

	embeddingstr := ""
	for _, v := range embedding {
		embeddingstr += strconv.FormatFloat(float64(v), 'f', 6, 64) + ","
	}
	embeddingstr = strings.TrimRight(embeddingstr, ",")

	err = DB.Exec(
		"INSERT INTO embeddings (memory_id, user_id, content, embedding) VALUES (?, ?::uuid, ?, string_to_array(?, ',')::float[])",
		memoryId,
		userId,
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

func GetMemoryIds(userId string) ([]MemoryRecord, error) {
	var results []MemoryRecord

	err := DB.Find(&results, "user_id = ?", userId).Error
	if err != nil {
		return nil, err
	}

	return results, nil
}

func MatchEmbeddings(memoryIds []string, userId string, embedding []float32) ([]MatchResult, error) {
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
