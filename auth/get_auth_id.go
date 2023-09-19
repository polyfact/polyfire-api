package auth

import (
	"net/http"

	router "github.com/julienschmidt/httprouter"
	db "github.com/polyfact/api/db"
	"github.com/polyfact/api/utils"
)

func GetAuthId(w http.ResponseWriter, r *http.Request, _ router.Params) {
	user_id := r.Context().Value(utils.ContextKeyUserID).(string)

	projectUser, _ := db.GetProjectUserByID(user_id)

	if projectUser == nil {
		_, _ = w.Write([]byte(user_id))
		return
	}

	_, _ = w.Write([]byte(projectUser.AuthID))
}
