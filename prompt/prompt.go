package prompt

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/polyfact/api/db"
	"github.com/polyfact/api/utils"
)

func GetPromptById(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	result, err := db.GetPromptById(id)
	if err != nil {
		utils.RespondError(w, "db_fetch_prompt_error")
		return
	}
	json.NewEncoder(w).Encode(result)
}

func GetPromptByName(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("name")
	result, err := db.GetPromptByName(name)
	if err != nil {
		utils.RespondError(w, "db_fetch_prompt_error")
		return
	}
	json.NewEncoder(w).Encode(result)
}

func GetAllPrompts(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	result, err := db.GetAllPrompts()
	if err != nil {
		utils.RespondError(w, "db_fetch_prompt_error")
		return
	}
	json.NewEncoder(w).Encode(result)
}

func CreatePrompt(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var input db.PromptInsert
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.RespondError(w, "decode_prompt_error")
		return
	}

	result, err := db.CreatePrompt(input)
	if err != nil {
		utils.RespondError(w, "db_insert_prompt_error")
		return
	}
	json.NewEncoder(w).Encode(result)
}

func UpdatePrompt(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	var input db.PromptUpdate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.RespondError(w, "decode_prompt_error")
		return
	}

	result, err := db.UpdatePrompt(id, input)
	if err != nil {
		utils.RespondError(w, "db_update_prompt_error")
		return
	}
	json.NewEncoder(w).Encode(result)
}

func DeletePrompt(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	if err := db.DeletePrompt(id); err != nil {
		utils.RespondError(w, "db_delete_prompt_error")
		return
	}
	w.WriteHeader(http.StatusOK)
}
