package auth

import (
	"net/http"

	router "github.com/julienschmidt/httprouter"
	database "github.com/polyfire/api/db"
	"github.com/polyfire/api/utils"
)

func GetAuthID(w http.ResponseWriter, r *http.Request, _ router.Params) {
	db := r.Context().Value(utils.ContextKeyDB).(database.Database)
	userID := r.Context().Value(utils.ContextKeyUserID).(string)

	projectUser, _ := db.GetProjectUserByID(userID)

	if projectUser == nil {
		_, _ = w.Write([]byte(userID))
		return
	}

	_, _ = w.Write([]byte(projectUser.AuthID))
}
