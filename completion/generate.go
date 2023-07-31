package completion

import (
	"encoding/json"
	"fmt"
	"net/http"

	router "github.com/julienschmidt/httprouter"
	db "github.com/polyfact/api/db"
	llm "github.com/polyfact/api/llm"
	memory "github.com/polyfact/api/memory"
	utils "github.com/polyfact/api/utils"
	llms "github.com/tmc/langchaingo/llms"
)

type GenerateRequestBody struct {
	Task     string    `json:"task"`
	MemoryId *string   `json:"memory_id,omitempty"`
	ChatId   *string   `json:"chat_id,omitempty"`
	Provider string    `json:"provider,omitempty"`
	Stop     *[]string `json:"stop,omitempty"`
	Stream   bool      `json:"stream,omitempty"`
}

func Generate(w http.ResponseWriter, r *http.Request, _ router.Params) {
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

		context_completion, err = utils.FillContext(results)

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
	if input.ChatId != nil && len(*input.ChatId) > 0 {
		chat, err := db.GetChatById(*input.ChatId)
		if err != nil {
			http.Error(w, "500 Internal server error", http.StatusInternalServerError)
			return
		}

		if chat == nil || chat.UserID != user_id {
			http.Error(w, "404 Not found", http.StatusNotFound)
			return
		}

		allHistory, err := db.GetChatMessages(user_id, *input.ChatId)
		if err != nil {
			http.Error(w, "500 Internal server error", http.StatusInternalServerError)
			return
		}

		chatHistory := utils.CutChatHistory(allHistory, 1000)

		var system_prompt string
		if chat.SystemPrompt == nil {
			system_prompt = ""
		} else {
			system_prompt = *(chat.SystemPrompt)
		}

		prompt := FormatPrompt(context_completion+"\n"+system_prompt, chatHistory, input.Task)

		fmt.Println(chat.ID)
		err = db.AddChatMessage(chat.ID, true, input.Task)
		if err != nil {
			http.Error(w, "500 Internal server error", http.StatusInternalServerError)
			fmt.Println(err)
			return
		}

		result, err = provider.Generate(prompt, &callback, &llms.CallOptions{StopWords: []string{"AI:", "Human:"}})

		if err != nil {
			http.Error(w, "500 Internal server error", http.StatusInternalServerError)
			return
		}

		err = db.AddChatMessage(chat.ID, false, result.Result)

	} else {
		prompt := context_completion + input.Task

		if input.Stop != nil {
			fmt.Println(*input.Stop)
			result, err = provider.Generate(prompt, &callback, &llms.CallOptions{StopWords: *input.Stop})
		} else {
			result, err = provider.Generate(prompt, &callback, nil)
		}
	}

	w.Header()["Content-Type"] = []string{"application/json"}

	if err != nil {
		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}
