package main

import (
	"log"
	"net/http"

	httprouter "github.com/julienschmidt/httprouter"

	auth "github.com/polyfact/api/auth"
	completion "github.com/polyfact/api/completion"
	imageGeneration "github.com/polyfact/api/image_generation"
	kv "github.com/polyfact/api/kv"
	memory "github.com/polyfact/api/memory"
	middlewares "github.com/polyfact/api/middlewares"
	"github.com/polyfact/api/prompt"
	transcription "github.com/polyfact/api/transcription"
)

func CORS(
	handler func(http.ResponseWriter, *http.Request, httprouter.Params),
) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		handler(w, r, params)
	}
}

func main() {
	log.Print("Starting the server on :8080")

	router := httprouter.New()

	router.POST("/generate", CORS(middlewares.Auth(completion.Generate)))

	router.GET("/chat/:id/history", CORS(middlewares.Auth(completion.GetChatHistory)))
	router.POST("/chats", CORS(middlewares.Auth(completion.CreateChat)))

	router.POST("/transcribe", CORS(middlewares.Auth(transcription.Transcribe)))

	router.GET("/image/generate", CORS(middlewares.Auth(imageGeneration.ImageGeneration)))

	router.GET("/memories", CORS(middlewares.Auth(memory.Get)))
	router.POST("/memory", CORS(middlewares.Auth(memory.Create)))
	router.PUT("/memory", CORS(middlewares.Auth(memory.Add)))
	router.GET("/project/:id/auth/token", CORS(auth.TokenExchangeHandler))
	router.GET("/usage", CORS(middlewares.Auth(auth.UserRateLimit)))

	router.GET("/kv", CORS(middlewares.Auth(kv.Get)))
	router.PUT("/kv", CORS(middlewares.Auth(kv.Set)))
	router.GET("/stream", CORS(middlewares.AuthStream(completion.Stream)))

	// Prompt Routes
	router.GET("/prompt/name/:name", CORS(middlewares.Auth(prompt.GetPromptByName)))
	router.GET("/prompt/id/:id", CORS(middlewares.Auth(prompt.GetPromptById)))
	router.GET("/prompts", CORS(middlewares.Auth(prompt.GetAllPrompts)))
	router.POST("/prompt", CORS(middlewares.Auth(prompt.CreatePrompt)))
	router.PUT("/prompt/:id", CORS(middlewares.Auth(prompt.UpdatePrompt)))
	router.DELETE("/prompt/:id", CORS(middlewares.Auth(prompt.DeletePrompt)))

	log.Fatal(http.ListenAndServe(":8080", router))
}
