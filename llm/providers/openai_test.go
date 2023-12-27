package providers

import (
	"context"
	"testing"

	"github.com/polyfire/api/utils"
)

func TestOpenAIProvider(t *testing.T) {
	ctx := utils.MockOpenAIServer(context.Background())
	result := NewOpenAIStreamProvider(ctx, "test-model").Generate("Test", nil, nil)

	str := ""

	for v := range result {
		str += v.Result
	}

	if str != "Test response" {
		t.Fatalf(`Generate("Test") should have returned "Test response" but returned "%s"`, str)
	}
}
