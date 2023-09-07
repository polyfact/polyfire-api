package auth

import (
	"net/http"

	router "github.com/julienschmidt/httprouter"
	db "github.com/polyfact/api/db"
)

func GetAuthId(w http.ResponseWriter, r *http.Request, _ router.Params) {
	user_id := r.Context().Value("user_id").(string)

	projectUser, _ := db.GetProjectUserByID(user_id)

	if projectUser == nil {
		w.Write([]byte(user_id))
	}

	w.Write([]byte(projectUser.AuthID))
}
