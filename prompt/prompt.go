package prompt

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/polyfact/api/db"
	"github.com/polyfact/api/utils"
)

func GetPromptById(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	result, err := db.GetPromptById(id)
	if err != nil {
		utils.RespondError(w, "db_fetch_prompt_error", err.Error())
		return
	}
	json.NewEncoder(w).Encode(result)
}

func GetPromptByName(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("name")
	result, err := db.GetPromptByName(name)
	if err != nil {
		utils.RespondError(w, "db_fetch_prompt_error", err.Error())
		return
	}
	json.NewEncoder(w).Encode(result)
}

func GetAllPrompts(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	queryParams := r.URL.Query()

	var filters db.SupabaseFilters

	for key, values := range queryParams {
		splitted := strings.Split(key, "_")
		if len(splitted) != 2 {
			errorMessage := key + " is not a valid filter format"
			utils.RespondError(w, "invalid_filter_format", errorMessage)
			return
		}

		column := splitted[0]
		operationStr := splitted[1]

		operation, err := db.StringToFilterOperation(operationStr)
		if err != nil {
			utils.RespondError(w, "invalid_filter_operation", err.Error())
			return
		}

		filter := db.SupabaseFilter{
			Column:    column,
			Value:     values[0],
			Operation: operation,
		}

		filters = append(filters, filter)
	}

	result, err := db.GetAllPrompts(filters)

	if err != nil {
		switch err.Error() {
		case "invalid_column":
			utils.RespondError(w, "invalid_column", err.Error())
			return
		case "invalid_operation":
			utils.RespondError(w, "invalid_length_value", err.Error())
			return
		default:
			utils.RespondError(w, "db_fetch_prompt_error", err.Error())
			return
		}
	}
	json.NewEncoder(w).Encode(result)
}

func CreatePrompt(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var input db.PromptInsert
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.RespondError(w, "decode_prompt_error", err.Error())
		return
	}

	result, err := db.CreatePrompt(input)

	if err != nil {
		utils.RespondError(w, "db_insert_prompt_error", err.Error())
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
		utils.RespondError(w, "db_update_prompt_error", err.Error())
		return
	}
	json.NewEncoder(w).Encode(result)
}

func DeletePrompt(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	if err := db.DeletePrompt(id); err != nil {
		utils.RespondError(w, "db_delete_prompt_error", err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}
