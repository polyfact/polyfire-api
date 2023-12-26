package completion

import (
	"context"
	"testing"

	database "github.com/polyfire/api/db"
	"github.com/polyfire/api/utils"
)

func mockCacheHit(_ string, _ string, _ string) (*database.CompletionCache, error) {
	return &database.CompletionCache{
		ID:        "00000000-0000-0000-0000-000000000000",
		Sha256sum: "0123456789abcdef",
		Result:    " john doe",
		Provider:  "test-provider",
		Model:     "test-model",
		Exact:     true,
	}, nil
}

func TestExactCacheHit(t *testing.T) {
	prompt := "My name is"

	ctx := context.WithValue(
		context.Background(),
		utils.ContextKeyDB,
		database.MockDatabase{MockGetExactCompletionCacheByHash: mockCacheHit},
	)

	result, err := CheckExactCache(ctx, prompt, "test-provider", "test-model")
	if err != nil {
		t.Fatalf(`CheckExactCache("My name is") returned an error: %v`, err)
	}

	resultStr := ""

	for v := range result {
		resultStr += v.Result
	}

	if resultStr != " john doe" {
		t.Fatalf(`CheckExactCache("My name is") should give " john doe". Result = "%v"`, resultStr)
	}
}

func mockCacheMiss(_ string, _ string, _ string) (*database.CompletionCache, error) {
	return nil, nil
}

func TestExactCacheMiss(t *testing.T) {
	prompt := "My name is"

	ctx := context.WithValue(
		context.Background(),
		utils.ContextKeyDB,
		database.MockDatabase{MockGetExactCompletionCacheByHash: mockCacheMiss},
	)

	result, err := CheckExactCache(ctx, prompt, "test-provider", "test-model")
	if err != nil {
		t.Fatalf(`CheckExactCache("My name is") returned an error: %v`, err)
	}

	if result != nil {
		t.Fatalf(`CheckExactCache("My name is") should return nil. result = %v`, result)
	}
}
