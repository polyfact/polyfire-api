package providers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func createTestClient() OpenAIStreamProvider {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(
			w,
			`data: {"id":"chatcmpl-8aKYlEvqL803VEkjJdxK3tsBAXU9T","object":"chat.completion.chunk","created":1703669531,"model":"gpt-3.5-turbo-0613","system_fingerprint":null,"choices":[{"index":0,"delta":{"role":"assistant","content":""},"logprobs":null,"finish_reason":null}]}

data: {"id":"chatcmpl-8aKYlEvqL803VEkjJdxK3tsBAXU9T","object":"chat.completion.chunk","created":1703669531,"model":"gpt-3.5-turbo-0613","system_fingerprint":null,"choices":[{"index":0,"delta":{"content":"Test"},"logprobs":null,"finish_reason":null}]}

data: {"id":"chatcmpl-8aKYlEvqL803VEkjJdxK3tsBAXU9T","object":"chat.completion.chunk","created":1703669531,"model":"gpt-3.5-turbo-0613","system_fingerprint":null,"choices":[{"index":0,"delta":{"content":" response"},"logprobs":null,"finish_reason":null}]}

data: {"id":"chatcmpl-8aKYlEvqL803VEkjJdxK3tsBAXU9T","object":"chat.completion.chunk","created":1703669531,"model":"gpt-3.5-turbo-0613","system_fingerprint":null,"choices":[{"index":0,"delta":{},"logprobs":null,"finish_reason":"stop"}]}

data: [DONE]`,
		)
	}))

	ctx := context.WithValue(
		context.Background(),
		ContextKeyHttpClient,
		server.Client(),
	)

	ctx = context.WithValue(ctx, ContextKeyBaseURL, server.URL)

	return NewOpenAIStreamProvider(ctx, "test-model")
}

func TestOpenAI(t *testing.T) {
	result := createTestClient().Generate("Test", nil, nil)

	str := ""

	for v := range result {
		str += v.Result
	}

	if str != "Test response" {
		t.Fatalf(`Generate("Test") should have returned "Test response" but returned "%s"`, str)
	}
}
