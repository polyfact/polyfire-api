package utils

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
)

func MockOpenAIServer(ctx context.Context) context.Context {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[INFO] Received request on mock OpenAI server url: %v\n", r.URL.Path)
		if r.URL.Path == "/chat/completions" {
			fmt.Fprintln(
				w,
				`data: {"id":"chatcmpl-mock","object":"chat.completion.chunk","created":1700000000,"model":"gpt-3.5-turbo-0613","system_fingerprint":null,"choices":[{"index":0,"delta":{"role":"assistant","content":""},"logprobs":null,"finish_reason":null}]}

data: {"id":"chatcmpl-mock","object":"chat.completion.chunk","created":1700000000,"model":"gpt-3.5-turbo-0613","system_fingerprint":null,"choices":[{"index":0,"delta":{"content":"Test"},"logprobs":null,"finish_reason":null}]}

data: {"id":"chatcmpl-mock","object":"chat.completion.chunk","created":1700000000,"model":"gpt-3.5-turbo-0613","system_fingerprint":null,"choices":[{"index":0,"delta":{"content":" response"},"logprobs":null,"finish_reason":null}]}

data: {"id":"chatcmpl-mock","object":"chat.completion.chunk","created":1700000000,"model":"gpt-3.5-turbo-0613","system_fingerprint":null,"choices":[{"index":0,"delta":{},"logprobs":null,"finish_reason":"stop"}]}

data: [DONE]`,
			)
		}
		if r.URL.Path == "/embeddings" {
			fmt.Fprintln(
				w,
				`{"object":"list","data":[{"object":"embedding","index":0,"embedding":[0.000000000,-1.000000000,1.000000000]}],"model":"text-embedding-ada-002-v2","usage":{"prompt_tokens":1,"total_tokens":1}}`,
			)
		}
	}))

	ctx = context.WithValue(
		ctx,
		ContextKeyHTTPClient,
		server.Client(),
	)

	ctx = context.WithValue(ctx, ContextKeyOpenAIBaseURL, server.URL)

	return ctx
}
