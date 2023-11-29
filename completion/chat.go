package completion

import (
	"encoding/json"
	"log"
	"net/http"

	router "github.com/julienschmidt/httprouter"
	"github.com/polyfire/api/db"
	"github.com/polyfire/api/llm/providers/options"
	"github.com/polyfire/api/utils"
)

func CreateChat(w http.ResponseWriter, r *http.Request, _ router.Params) {
	userID := r.Context().Value(utils.ContextKeyUserID).(string)
	record := r.Context().Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)

	var requestBody struct {
		SystemPrompt   *string `json:"system_prompt"`
		SystemPromptID *string `json:"system_prompt_id"`
		Name           *string `json:"name"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestBody); err != nil {
		utils.RespondError(w, record, "decode_error")
		return
	}

	systemPromptId, err := db.RetrieveSystemPromptID(requestBody.SystemPromptID)

	if err != nil {
		utils.RespondError(w, record, "error_retrieving_system_prompt_id")
		return
	}

	chat, err := db.CreateChat(userID, requestBody.SystemPrompt, systemPromptId, requestBody.Name)
	if err != nil {
		log.Printf("Error creating chat for user %s : %v", userID, err)
		utils.RespondError(w, record, "error_create_chat", err.Error())
		return
	}

	response, _ := json.Marshal(&chat)
	record(string(response))

	_ = json.NewEncoder(w).Encode(chat)
}

func UpdateChat(w http.ResponseWriter, r *http.Request, ps router.Params) {
	id := ps.ByName("id")
	userID := r.Context().Value(utils.ContextKeyUserID).(string)
	record := r.Context().Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)

	var requestBody struct {
		Name string `json:"name"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestBody); err != nil {
		utils.RespondError(w, record, "decode_error")
		return
	}

	chat, err := db.UpdateChat(userID, id, requestBody.Name)
	if err != nil {
		utils.RespondError(w, record, "error_update_chat", err.Error())
		return
	}

	response, _ := json.Marshal(&chat)
	record(string(response))

	_ = json.NewEncoder(w).Encode(chat)
}

func DeleteChat(w http.ResponseWriter, r *http.Request, ps router.Params) {
	id := ps.ByName("id")
	userID := r.Context().Value(utils.ContextKeyUserID).(string)
	record := r.Context().Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)

	err := db.DeleteChat(userID, id)
	if err != nil {
		utils.RespondError(w, record, "error_delete_chat", err.Error())
		return
	}

	_, _ = w.Write([]byte("{\"success\":true}"))
}

func ListChat(w http.ResponseWriter, r *http.Request, _ router.Params) {
	userID := r.Context().Value(utils.ContextKeyUserID).(string)
	record := r.Context().Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)

	chats, err := db.ListChats(userID)
	if err != nil {
		utils.RespondError(w, record, "error_list_chat", err.Error())
		return
	}

	response, _ := json.Marshal(&chats)
	record(string(response))

	_ = json.NewEncoder(w).Encode(chats)
}

func GetChatHistory(w http.ResponseWriter, r *http.Request, ps router.Params) {
	id := ps.ByName("id")
	userID := r.Context().Value(utils.ContextKeyUserID).(string)
	record := r.Context().Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)

	messages, err := db.GetChatMessages(userID, id)
	if err != nil {
		utils.RespondError(w, record, "error_chat_history")
		return
	}

	response, _ := json.Marshal(&messages)
	record(string(response))

	_ = json.NewEncoder(w).Encode(messages)
}

func AddToChatHistory(
	userID string,
	task string,
	chatID string,
	callback options.ProviderCallback,
	opts *options.ProviderOptions,
) error {
	log.Println("GetChatByID")
	chat, err := db.GetChatByID(chatID)
	if err != nil {
		return ErrInternalServerError
	}

	if chat == nil || chat.UserID != userID {
		return ErrNotFound
	}

	oldCallback := *callback
	*callback = func(providerName string, modelName string, inputCount int, outputCount int, completion string, credit *int) {
		if oldCallback != nil {
			log.Println("Old callback")
			oldCallback(providerName, modelName, inputCount, outputCount, completion, credit)
		}

		log.Println("Add Chat Message")
		err = db.AddChatMessage(chat.ID, true, task)
		if err != nil {
			log.Printf("Error adding chat message for user %s : %v", userID, err)
		}
		log.Println("Add Chat Message Callback")
		_ = db.AddChatMessage(chat.ID, false, completion)
	}

	opts.StopWords = &[]string{"User:", "You:"}

	return nil
}
