package main

import (
	"log"
	"net/http"

	httprouter "github.com/julienschmidt/httprouter"

	auth "github.com/polyfact/api/auth"
	completion "github.com/polyfact/api/completion"
	db "github.com/polyfact/api/db"
	imageGeneration "github.com/polyfact/api/image_generation"
	kv "github.com/polyfact/api/kv"
	memory "github.com/polyfact/api/memory"
	middlewares "github.com/polyfact/api/middlewares"
	prompt "github.com/polyfact/api/prompt"
	transcription "github.com/polyfact/api/transcription"
)

type CORSRouter struct {
	Router *httprouter.Router
}

func (h CORSRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")

	middlewares.AddRecord(r)
	defer middlewares.RecoverFromPanic(w, r)

	h.Router.ServeHTTP(w, r)
}

func GlobalMiddleware(router *httprouter.Router) http.Handler {
	return &CORSRouter{Router: router}
}

func main() {
	log.Print("Starting the server on :8080")

	db.InitDB()

	router := httprouter.New()

	// Auth Routes
	router.GET("/project/:id/auth/token", auth.ManagedSupabaseTokenExchangeHandler)
	router.GET("/project/:id/auth/firebase", auth.ExternalFirebaseTokenExchangeHandler)
	router.GET("/project/:id/auth/custom/token", auth.ExternalCustomTokenExchangeHandler)
	router.GET("/auth/id", middlewares.Auth(auth.GetAuthId))
	router.GET("/usage", middlewares.Auth(auth.UserRateLimit))

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

	// KV Routes
	router.GET("/kv", middlewares.Auth(kv.Get))
	router.GET("/kvs", middlewares.Auth(kv.List))
	router.PUT("/kv", middlewares.Auth(kv.Set))
	router.DELETE("/kv", middlewares.Auth(kv.Delete))

	// Prompt Routes
	router.GET("/prompt/name/:name", middlewares.Auth(prompt.GetPromptByName))
	router.GET("/prompt/id/:id", middlewares.Auth(prompt.GetPromptById))
	router.GET("/prompts", middlewares.Auth(prompt.GetAllPrompts))
	router.POST("/prompt", middlewares.Auth(prompt.CreatePrompt))
	router.PUT("/prompt/:id", middlewares.Auth(prompt.UpdatePrompt))
	router.DELETE("/prompt/:id", middlewares.Auth(prompt.DeletePrompt))

	// Prompt Like Routes
	router.POST("/prompt/like/:id", middlewares.Auth(prompt.HandlePromptLike))

	log.Fatal(http.ListenAndServe(":8080", GlobalMiddleware(router)))
}
