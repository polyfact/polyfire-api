package db

import (
	"strconv"
	"strings"
)

type CompletionCache struct {
	ID        string     `json:"id"`
	Input     FloatArray `json:"input"`
	Result    string     `json:"result"`
	Provider  string     `json:"provider"`
	Model     string     `json:"model"`
	CreatedAt string     `json:"created_at"`
}

func (CompletionCache) TableName() string {
	return "completion_cache"
}

func GetCompletionCache(id string) (*CompletionCache, error) {
	var cache []CompletionCache
	err := DB.Find(&cache, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	if len(cache) == 0 {
		return nil, nil
	}

	return &cache[0], nil
}

func GetCompletionCacheByInput(provider string, model string, input []float32) (*CompletionCache, error) {
	embeddingstr := "["
	for _, v := range input {
		embeddingstr += strconv.FormatFloat(float64(v), 'f', 6, 64) + ","
	}
	embeddingstr = strings.TrimRight(embeddingstr, ",") + "]"

	var cache []CompletionCache
	err := DB.Find(
		&cache,
		"provider = ? AND model = ? AND input <-> ? < 0.15 ORDER BY input <-> ? ASC",
		provider,
		model,
		embeddingstr,
		embeddingstr,
	).Error
	if err != nil {
		return nil, err
	}

	if len(cache) == 0 {
		return nil, nil
	}

	return &cache[0], nil
}

func AddCompletionCache(input []float32, result string, provider string, model string) error {
	embeddingstr := ""
	for _, v := range input {
		embeddingstr += strconv.FormatFloat(float64(v), 'f', 6, 64) + ","
	}
	embeddingstr = strings.TrimRight(embeddingstr, ",") + ""

	err := DB.Exec(
		"INSERT INTO completion_cache (result, provider, model, input) VALUES (?, ?, ?, string_to_array(?, ',')::float[])",
		result, provider, model, embeddingstr).Error

	return err
}
