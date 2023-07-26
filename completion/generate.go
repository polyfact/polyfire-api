package completion

import (
	"encoding/json"
	"fmt"
	"net/http"

	db "github.com/polyfact/api/db"
	llm "github.com/polyfact/api/llm"
	memory "github.com/polyfact/api/memory"
	helpers "github.com/polyfact/api/utils"
)

type GenerateRequestBody struct {
	Task     string  `json:"task"`
	MemoryId *string `json:"memory_id,omitempty"`
	Provider string  `json:"provider,omitempty"`
}

func Generate(w http.ResponseWriter, r *http.Request) {
	user_id := r.Context().Value("user_id").(string)

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

	context_completion := ""

	if input.MemoryId != nil && len(*input.MemoryId) > 0 {
		results, err := memory.Embedder(user_id, *input.MemoryId, input.Task)
		if err != nil {
			http.Error(w, "500 Internal server error", http.StatusInternalServerError)
			return
		}

		context_completion, err = helpers.FillContext(results)

		if err != nil {
			http.Error(w, "500 Internal server error", http.StatusInternalServerError)
			return
		}

	}

	callback := func(model_name string, input_count int, output_count int) {
		db.LogRequests(user_id, model_name, input_count, output_count, "completion")
	}

	if input.Provider == "" {
		input.Provider = "openai"
	}

	provider, err := llm.NewLLMProvider(input.Provider)
	if err == llm.ErrUnknownModel {
		http.Error(w, "400 Unknown model provider", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	var result llm.Result
	var prompt string = context_completion + input.Task

	result, err = provider.Generate(prompt, &callback)

	w.Header()["Content-Type"] = []string{"application/json"}

	if err != nil {
		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}
