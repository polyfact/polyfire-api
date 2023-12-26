package completion

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	router "github.com/julienschmidt/httprouter"
	database "github.com/polyfire/api/db"
	"github.com/polyfire/api/llm/providers/options"
	"github.com/polyfire/api/utils"
)

func CreateChat(w http.ResponseWriter, r *http.Request, _ router.Params) {
	db := r.Context().Value(utils.ContextKeyDB).(database.DB)
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

	systemPromptID, err := db.RetrieveSystemPromptID(requestBody.SystemPromptID)
	if err != nil {
		utils.RespondError(w, record, "error_retrieving_system_prompt_id")
		return
	}

	chat, err := db.CreateChat(userID, requestBody.SystemPrompt, systemPromptID, requestBody.Name)
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
	db := r.Context().Value(utils.ContextKeyDB).(database.DB)
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
	db := r.Context().Value(utils.ContextKeyDB).(database.DB)
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
	db := r.Context().Value(utils.ContextKeyDB).(database.DB)
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
	db := r.Context().Value(utils.ContextKeyDB).(database.DB)
	id := ps.ByName("id")
	orderByParam := r.URL.Query().Get("orderBy")
	limitParam := r.URL.Query().Get("limit")
	offsetParam := r.URL.Query().Get("offset")

	userID := r.Context().Value(utils.ContextKeyUserID).(string)
	record := r.Context().Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)

	orderByDESC := true
	if strings.ToLower(orderByParam) == "asc" || orderByParam == "1" {
		orderByDESC = false
	}

	limit := 20
	if val, err := strconv.Atoi(limitParam); err == nil {
		limit = val
	}

	offset, _ := strconv.Atoi(offsetParam)

	messages, err := db.GetChatMessages(userID, id, orderByDESC, limit, offset)
	if err != nil {
		utils.RespondError(w, record, "error_chat_history")
		return
	}

	response, _ := json.Marshal(&messages)
	record(string(response))

	_ = json.NewEncoder(w).Encode(messages)
}

func AddToChatHistory(
	ctx context.Context,
	userID string,
	task string,
	chatID string,
	callback options.ProviderCallback,
	opts *options.ProviderOptions,
) error {
	db := ctx.Value(utils.ContextKeyDB).(database.DB)
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
