package db

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"strconv"
	"strings"
)

type CompletionCache struct {
	ID        string     `json:"id"`
	Sha256sum string     `json:"sha256sum"`
	Input     FloatArray `json:"input"`
	Result    string     `json:"result"`
	Provider  string     `json:"provider"`
	Model     string     `json:"model"`
	Exact     bool       `json:"exact"`
	CreatedAt string     `json:"created_at"`
}

func (CompletionCache) TableName() string {
	return "completion_cache"
}

func (db DB) GetCompletionCache(id string) (*CompletionCache, error) {
	var cache []CompletionCache
	err := db.sql.Find(&cache, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	if len(cache) == 0 {
		return nil, nil
	}

	return &cache[0], nil
}

func (db DB) GetCompletionCacheByInput(provider string, model string, input []float32) (*CompletionCache, error) {
	embeddingstr := "["
	for _, v := range input {
		embeddingstr += strconv.FormatFloat(float64(v), 'f', 6, 64) + ","
	}
	embeddingstr = strings.TrimRight(embeddingstr, ",") + "]"

	var cache []CompletionCache
	err := db.sql.Find(
		&cache,
		"exact = false AND provider = ? AND model = ? AND input <-> ? < 0.15 ORDER BY input <-> ? ASC",
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

func (db DB) AddCompletionCache(
	input []float32,
	prompt string,
	result string,
	provider string,
	model string,
	exact bool,
) error {
	embeddingstr := ""
	for _, v := range input {
		embeddingstr += strconv.FormatFloat(float64(v), 'f', 6, 64) + ","
	}
	embeddingstr = strings.TrimRight(embeddingstr, ",") + ""

	sha256sum := sha256.Sum256([]byte(prompt))
	sha256sumHex := hex.EncodeToString(sha256sum[:])

	err := db.sql.Exec(
		"INSERT INTO completion_cache (result, provider, model, input, exact, sha256sum) VALUES (@result, @provider, @model, CASE WHEN @input = '' THEN NULL ELSE string_to_array(@input, ',')::float[] END, @exact, @sha256sum) ON CONFLICT (sha256sum) DO NOTHING;",
		sql.Named("result", result),
		sql.Named("provider", provider),
		sql.Named("model", model),
		sql.Named("input", embeddingstr),
		sql.Named("exact", exact),
		sql.Named("sha256sum", sha256sumHex),
	).Error

	return err
}

func (db DB) GetExactCompletionCacheByHash(provider string, model string, input string) (*CompletionCache, error) {
	sha256sum := sha256.Sum256([]byte(input))
	sha256sumHex := hex.EncodeToString(sha256sum[:])

	var cache []CompletionCache
	err := db.sql.Find(
		&cache,
		"provider = ? AND model = ? AND sha256sum = ?",
		provider,
		model,
		sha256sumHex,
	).Error
	if err != nil {
		return nil, err
	}

	if len(cache) == 0 {
		return nil, nil
	}

	return &cache[0], nil
}
