package utils

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
)

func MockOpenAIServer(ctx context.Context) context.Context {
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

	ctx = context.WithValue(
		ctx,
		ContextKeyHTTPClient,
		server.Client(),
	)

	ctx = context.WithValue(ctx, ContextKeyOpenAIBaseURL, server.URL)

	return ctx
}
