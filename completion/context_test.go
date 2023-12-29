package completion

import (
	"context"
	"strings"
	"testing"

	database "github.com/polyfire/api/db"
	"github.com/polyfire/api/utils"
)

func mockGetExistingEmbeddingFromContent(_ string) (*[]float32, error) {
	return nil, nil
}

func mockMatchEmbeddings(_ []string, _ string, _ []float32) ([]database.MatchResult, error) {
	result := database.MatchResult{
		ID:         "00000000-0000-0000-0000-000000000000",
		Content:    "banana42",
		Similarity: 0.9,
	}

	return []database.MatchResult{result}, nil
}

func TestContextStringMemory(t *testing.T) {
	utils.SetLogLevel("WARN")
	ctx := context.Background()

	ctx = utils.MockOpenAIServer(ctx)

	userID := "00000000-0000-0000-0000-000000000000"

	ctx = context.WithValue(ctx, utils.ContextKeyUserID, userID)

	ctx = context.WithValue(ctx, utils.ContextKeyEventID, "00000000-0000-0000-0000-000000000000")

	ctx = context.WithValue(
		ctx,
		utils.ContextKeyDB,
		database.MockDatabase{
			MockGetExistingEmbeddingFromContent: mockGetExistingEmbeddingFromContent,
			MockLogRequests:                     mockLogRequests,
			MockMatchEmbeddings:                 mockMatchEmbeddings,
		},
	)

	reqBody := GenerateRequestBody{
		Task:     "Test",
		MemoryID: "11100000-0000-0000-0000-000000000000",
	}

	result, _, err := GetContextString(ctx, userID, reqBody, nil, nil)
	if err != nil {
		t.Fatalf(`GetContextString returned an error %v`, err)
	}

	if !strings.Contains(result, "banana42") {
		t.Fatalf(`GetContextString doesn't contains "banana42". ContextString: "%s"`, result)
	}
}
