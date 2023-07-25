package main

import (
	"log"
	"net/http"

	completion "github.com/polyfact/api/completion"
	memory "github.com/polyfact/api/memory"
	middlewares "github.com/polyfact/api/middlewares"
	transcription "github.com/polyfact/api/transcription"
)

func main() {
	log.Print("Starting the server on :8080")
	http.HandleFunc("/generate", middlewares.Auth(completion.Generate))
	http.HandleFunc("/memory", middlewares.Auth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			memory.Create(w, r)
		} else if r.Method == http.MethodPut {
			memory.Add(w, r)
		} else {
			http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	http.HandleFunc("/memories", middlewares.Auth(memory.Get))
	http.HandleFunc("/transcribe", middlewares.Auth(transcription.Transcribe))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
