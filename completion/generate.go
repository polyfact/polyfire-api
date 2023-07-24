package completion

import (
	"encoding/json"
	"net/http"

	db "github.com/polyfact/api/db"
	llm "github.com/polyfact/api/llm"
	memory "github.com/polyfact/api/memory"
	helpers "github.com/polyfact/api/utils"
)

type GenerateRequestBody struct {
	Task      string  `json:"task"`
	Memory_id *string `json:"memory_id,omitempty"`
}

func Generate(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("user_id").(string)

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

	contextCompletion := ""

	if input.Memory_id != nil {
		results, err := memory.Embedder(userId, *input.Memory_id, input.Task)

		if err != nil {
			http.Error(w, "500 Internal server error", http.StatusInternalServerError)
			return
		}

		contextCompletion, err = helpers.FillContext(results)

		if err != nil {
			http.Error(w, "500 Internal server error", http.StatusInternalServerError)
			return
		}

	}

	callback := func(model_name string, input_count int, output_count int) {
		db.LogRequests(userId, model_name, input_count, output_count, "completion")
	}

	var result llm.Result
	var prompt string = contextCompletion + input.Task

	result, err = llm.Generate(prompt, &callback)

	w.Header()["Content-Type"] = []string{"application/json"}

	if err != nil {
		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}
