package completion

import (
	"encoding/json"
	"log"
	"net/http"

	router "github.com/julienschmidt/httprouter"
	db "github.com/polyfact/api/db"
	utils "github.com/polyfact/api/utils"
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

func CreateChat(w http.ResponseWriter, r *http.Request, _ router.Params) {
	user_id := r.Context().Value("user_id").(string)

	var requestBody struct {
		SystemPrompt *string `json:"system_prompt"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestBody); err != nil {
		utils.RespondError(w, "decode_error")
		return
	}

	if r.Method != "POST" {
		utils.RespondError(w, "only_post_method_allowed")
		return
	}

	chat, err := db.CreateChat(user_id, requestBody.SystemPrompt)
	if err != nil {
		log.Printf("Error creating chat for user %s with system prompt %v: %v", user_id, requestBody.SystemPrompt, err)
		utils.RespondError(w, "error_create_chat")
		return
	}

	json.NewEncoder(w).Encode(chat)
}

func GetChatHistory(w http.ResponseWriter, r *http.Request, ps router.Params) {
	id := ps.ByName("id")
	user_id := r.Context().Value("user_id").(string)

	messages, err := db.GetChatMessages(user_id, id)
	if err != nil {
		utils.RespondError(w, "error_chat_history")
		return
	}

	json.NewEncoder(w).Encode(messages)
}
