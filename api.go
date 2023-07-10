package main

import (
	"encoding/json"
	"log"
	"net/http"

	llm "github.com/polyfact/api/llm"
)

type GenerateRequestBody struct {
	Task       string `json:"task"`
	ReturnType any    `json:"return_type"`
}

func generate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if len(r.Header["Content-Type"]) == 0 || r.Header["Content-Type"][0] != "application/json" {
		http.Error(w, "400 bad request", http.StatusBadRequest)
		return
	}

	var input GenerateRequestBody

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "400 bad request", http.StatusBadRequest)
		return
	}

	result, err := llm.Generate(input.ReturnType, input.Task)

	if err != nil {
		http.Error(w, "500 task failed to execute", http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(result)
}

func main() {
	http.HandleFunc("/generate", generate)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
