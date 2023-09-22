package prompt

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/polyfact/api/db"
	"github.com/polyfact/api/utils"
)

const (
	DBFetchError      = "db_fetch_prompt_like_error"
	DBAddLikeError    = "db_add_prompt_like_error"
	DBRemoveLikeError = "db_remove_prompt_like_error"
)

func HandlePromptLike(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	record := r.Context().Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)
	userID := r.Context().Value(utils.ContextKeyUserID).(string)

	input := db.PromptLikeInput{
		UserId:   userID,
		PromptId: ps.ByName("id"),
	}

	prompts, err := db.GetPromptLikeByUserId(input)
	if err != nil {
		utils.RespondError(w, record, DBFetchError, err.Error())
		return
	}

	var data db.PromptLikeOutput

	if len(prompts) == 0 {
		if _, err := db.AddPromptLike(input); err != nil {
			utils.RespondError(w, record, DBAddLikeError, err.Error())
			return
		}
		data = db.PromptLikeOutput{
			UserId:   input.UserId,
			PromptId: input.PromptId,
			Like:     true,
		}
	} else {
		if _, err := db.RemovePromptLike(input); err != nil {
			utils.RespondError(w, record, DBRemoveLikeError, err.Error())
			return
		}
		data = db.PromptLikeOutput{
			UserId:   input.UserId,
			PromptId: input.PromptId,
			Like:     false,
		}
	}

	response, _ := json.Marshal(data)
	record(string(response))
	_ = json.NewEncoder(w).Encode(data)
}
