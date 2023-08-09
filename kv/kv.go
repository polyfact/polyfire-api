package kv

import (
	"encoding/json"
	"net/http"

	router "github.com/julienschmidt/httprouter"
	db "github.com/polyfact/api/db"
	"github.com/polyfact/api/utils"
)

func Get(w http.ResponseWriter, r *http.Request, _ router.Params) {
	user_id := r.Context().Value("user_id").(string)

	id := r.URL.Query().Get("key")

	if id == "" {
		utils.RespondError(w, "missing_id")
		return
	}

	store, _ := db.GetKV(user_id, id)

	if store == nil || store.Value == "" {
		utils.RespondError(w, "data_not_found")
		return
	}

	w.Write([]byte(store.Value))
}

func Set(w http.ResponseWriter, r *http.Request, _ router.Params) {
	user_id := r.Context().Value("user_id").(string)

	var data struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		utils.RespondError(w, "decode_error")
		return
	}

	err = db.SetKV(user_id, data.Key, data.Value)
	if err != nil {
		utils.RespondError(w, "database_error")
		return
	}

	w.WriteHeader(http.StatusCreated)
}
