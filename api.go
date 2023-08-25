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

type CORSHandler struct {
	Handler http.Handler
}

func (h CORSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	h.Handler.ServeHTTP(w, r)
}

func CORS(
	handler http.Handler,
) http.Handler {
	return CORSHandler{
		Handler: handler,
	}
}

func main() {
	log.Print("Starting the server on :8080")

	router := httprouter.New()

	router.POST("/generate", middlewares.Auth(completion.Generate))

	router.GET("/chat/:id/history", middlewares.Auth(completion.GetChatHistory))
	router.POST("/chats", middlewares.Auth(completion.CreateChat))

	router.POST("/transcribe", middlewares.Auth(transcription.Transcribe))

	router.GET("/image/generate", middlewares.Auth(imageGeneration.ImageGeneration))

	router.GET("/memories", middlewares.Auth(memory.Get))
	router.POST("/memory", middlewares.Auth(memory.Create))
	router.PUT("/memory", middlewares.Auth(memory.Add))
	router.GET("/project/:id/auth/token", auth.TokenExchangeHandler)
	router.GET("/usage", middlewares.Auth(auth.UserRateLimit))

	router.GET("/kv", middlewares.Auth(kv.Get))
	router.PUT("/kv", middlewares.Auth(kv.Set))
	router.GET("/stream", middlewares.AuthStream(completion.Stream))

	// Prompt Routes
	router.GET("/prompt/name/:name", middlewares.Auth(prompt.GetPromptByName))
	router.GET("/prompt/id/:id", middlewares.Auth(prompt.GetPromptById))
	router.GET("/prompts", middlewares.Auth(prompt.GetAllPrompts))
	router.POST("/prompt", middlewares.Auth(prompt.CreatePrompt))
	router.PUT("/prompt/:id", middlewares.Auth(prompt.UpdatePrompt))
	router.DELETE("/prompt/:id", middlewares.Auth(prompt.DeletePrompt))

	log.Fatal(http.ListenAndServe(":8080", CORS(router)))
}
