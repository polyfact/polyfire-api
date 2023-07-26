package completion

import (
	"encoding/json"
	"net/http"

	db "github.com/polyfact/api/db"
)

func FormatPrompt(systemPrompt string, chatHistory []db.ChatMessage, userPrompt string) string {
	res := systemPrompt

	for i := len(chatHistory) - 1; i >= 0; i-- {
		if chatHistory[i].IsUserMessage {
			res += "\nHuman: " + chatHistory[i].Content
		} else {
			res += "\nAI: " + chatHistory[i].Content
		}
	}

	res += "\nHuman: " + userPrompt + "\nAI: "

	return res
}

func CreateChat(w http.ResponseWriter, r *http.Request) {
	user_id := r.Context().Value("user_id").(string)

	if r.Method != "POST" {
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		return
	}

	chat, err := db.CreateChat(user_id)
	if err != nil {
		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(chat)
}
