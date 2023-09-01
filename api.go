package main

import (
	"bytes"
	"context"
	"io/ioutil"
	"log"
	"net/http"

	httprouter "github.com/julienschmidt/httprouter"

	auth "github.com/polyfact/api/auth"
	completion "github.com/polyfact/api/completion"
	imageGeneration "github.com/polyfact/api/image_generation"
	kv "github.com/polyfact/api/kv"
	memory "github.com/polyfact/api/memory"
	middlewares "github.com/polyfact/api/middlewares"
	posthog "github.com/polyfact/api/posthog"
	"github.com/polyfact/api/prompt"
	transcription "github.com/polyfact/api/transcription"
)

type CORSRouter struct {
	Router *httprouter.Router
}

func (h CORSRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	recordEventRequest := func(request string, response string, userID string) {
		properties := make(map[string]string)
		properties["path"] = string(r.URL.Path)
		properties["requestBody"] = request
		properties["responseBody"] = response
		posthog.Event("API Request", userID, properties)
	}

	buf, _ := ioutil.ReadAll(r.Body) // handle the error
	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))

	r.Body = rdr1

	recordEventWithUserID := func(response string, userID string) {
		recordEventRequest(string(buf), response, "")
	}

	recordEvent := func(response string) {
		recordEventWithUserID(response, "")
	}

	newCtx := context.WithValue(r.Context(), "recordEvent", recordEvent)
	newCtx = context.WithValue(newCtx, "recordEventRequest", recordEventRequest)
	newCtx = context.WithValue(newCtx, "recordEventWithUserID", recordEventWithUserID)

	defer middlewares.RecoverFromPanic(w, recordEvent)

	h.Router.ServeHTTP(w, r.WithContext(newCtx))
}

func GlobalMiddleware(router *httprouter.Router) http.Handler {
	return &CORSRouter{Router: router}
}

func main() {
	log.Print("Starting the server on :8080")

	router := httprouter.New()

	// Completion Routes
	router.POST("/generate", middlewares.Auth(completion.Generate))
	router.GET("/chat/:id/history", middlewares.Auth(completion.GetChatHistory))
	router.POST("/chats", middlewares.Auth(completion.CreateChat))
	router.GET("/stream", middlewares.AuthStream(completion.Stream))

	// Transcription Routes
	router.POST("/transcribe", middlewares.Auth(transcription.Transcribe))

	// Image Generation Routes
	router.GET("/image/generate", middlewares.Auth(imageGeneration.ImageGeneration))

	// Memory Routes
	router.GET("/memories", middlewares.Auth(memory.Get))
	router.POST("/memory", middlewares.Auth(memory.Create))
	router.PUT("/memory", middlewares.Auth(memory.Add))

	// Auth Routes
	router.GET("/project/:id/auth/token", auth.TokenExchangeHandler)
	router.GET("/usage", middlewares.Auth(auth.UserRateLimit))

	// KV Routes
	router.GET("/kv", middlewares.Auth(kv.Get))
	router.PUT("/kv", middlewares.Auth(kv.Set))

	// Prompt Routes
	router.GET("/prompt/name/:name", middlewares.Auth(prompt.GetPromptByName))
	router.GET("/prompt/id/:id", middlewares.Auth(prompt.GetPromptById))
	router.GET("/prompts", middlewares.Auth(prompt.GetAllPrompts))
	router.POST("/prompt", middlewares.Auth(prompt.CreatePrompt))
	router.PUT("/prompt/:id", middlewares.Auth(prompt.UpdatePrompt))
	router.DELETE("/prompt/:id", middlewares.Auth(prompt.DeletePrompt))

	log.Fatal(http.ListenAndServe(":8080", GlobalMiddleware(router)))
}
