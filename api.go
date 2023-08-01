package main

import (
	"log"
	"net/http"

	httprouter "github.com/julienschmidt/httprouter"

	completion "github.com/polyfact/api/completion"
	kv "github.com/polyfact/api/kv"
	memory "github.com/polyfact/api/memory"
	middlewares "github.com/polyfact/api/middlewares"
	transcription "github.com/polyfact/api/transcription"
)

func main() {
	log.Print("Starting the server on :8080")

	router := httprouter.New()

	router.POST("/generate", middlewares.Auth(completion.Generate))

	router.GET("/chat/:id/history", middlewares.Auth(completion.GetChatHistory))
	router.POST("/chats", middlewares.Auth(completion.CreateChat))

	router.POST("/transcribe", middlewares.Auth(transcription.Transcribe))

	router.GET("/memories", middlewares.Auth(memory.Get))
	router.POST("/memory", middlewares.Auth(memory.Create))
	router.PUT("/memory", middlewares.Auth(memory.Add))

	router.GET("/kv", middlewares.Auth(kv.Get))
	router.PUT("/kv", middlewares.Auth(kv.Set))
	router.GET("/stream", middlewares.Auth(completion.Stream))
	log.Fatal(http.ListenAndServe(":8080", router))
}
