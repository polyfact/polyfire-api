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

		if r.URL.Path == "/images/generations" {
			fmt.Fprintln(
				w,
				`{
				  "created": 1704289612,
				  "data": [
				    {
							"b64_json": "iVBORw0KGgoAAAANSUhEUgAAAAoAAAAKCAIAAAACUFjqAAABhWlDQ1BJQ0MgcHJvZmlsZQAAKJF9kT1Iw1AUhU9TRZEWBztIcchQBcEuKuJYq1CECqFWaNXB5KV/0KQhSXFxFFwLDv4sVh1cnHV1cBUEwR8QZwcnRRcp8b6k0CLGC4/3cd49h/fuA4RmlWlWTwLQdNvMpJJiLr8q9r0igCDCGEdUZpYxJ0lp+NbXPXVT3cV5ln/fnxVWCxYDAiJxghmmTbxBPLNpG5z3iSOsLKvE58QTJl2Q+JHrisdvnEsuCzwzYmYz88QRYrHUxUoXs7KpEU8Tx1RNp3wh57HKeYuzVq2z9j35C0MFfWWZ67RGkMIiliBBhII6KqjCRpx2nRQLGTpP+vijrl8il0KuChg5FlCDBtn1g//B79laxalJLymUBHpfHOdjFOjbBVoNx/k+dpzWCRB8Bq70jr/WBGY/SW90tNgRMLgNXFx3NGUPuNwBhp8M2ZRdKUhLKBaB9zP6pjwwdAsMrHlza5/j9AHI0qzSN8DBITBWoux1n3f3d8/t3572/H4Ab09ypVhh4DcAAAAJcEhZcwAALiMAAC4jAXilP3YAAAAHdElNRQfoAQMNKjBYfNrGAAAAGXRFWHRDb21tZW50AENyZWF0ZWQgd2l0aCBHSU1QV4EOFwAAAFZJREFUGNOFkLENACEMAw8Wos0AjJAxaekyRJjIX3z3euDkzrJku0gCWIsIxgBwp3daA5CkOQVfzSkJZf54rzIrEeyIQGbbtFnlSMV9a7pfql2GlfMtD9UkXRG5MHU5AAAAAElFTkSuQmCC",
				      "revised_prompt": "A simple red circle. The round shape is perfect and unbroken, with a uniformly filled-in deep crimson hue. The background is a neutral white, providing stark contrast to the vibrant red of the circle.",
				      "url": "https://oaidalleapiprodscus.blob.core.windows.net/private/org-wEJEnFznbktHKebSaZjm6Vx8/user-huXdfKX2z4h2NRhUiYfHDUQI/img-yh3k1kb1LOISYH9C8si1g3Hd.png?st=2024-01-03T12%3A46%3A52Z&se=2024-01-03T14%3A46%3A52Z&sp=r&sv=2021-08-06&sr=b&rscd=inline&rsct=image/png&skoid=6aaadede-4fb3-4698-a8f6-684d7786b067&sktid=a48cca56-e6da-484e-a814-9c849652bcb3&skt=2024-01-02T15%3A12%3A23Z&ske=2024-01-03T15%3A12%3A23Z&sks=b&skv=2021-08-06&sig=dVVRfUpQAy/kaIrOYY8RPpYhI7Tx72iexR9phanispo%3D"
				    }
				  ]
				}`,
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
