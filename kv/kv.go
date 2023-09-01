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
	record := r.Context().Value("recordEvent").(func(response string))

	id := r.URL.Query().Get("key")

	if id == "" {
		utils.RespondError(w, record, "missing_id")
		return
	}

	store, _ := db.GetKV(user_id, id)

	if store == nil || store.Value == "" {
		utils.RespondError(w, record, "data_not_found")
		return
	}

	record(store.Value)

	w.Write([]byte(store.Value))
}

func Set(w http.ResponseWriter, r *http.Request, _ router.Params) {
	user_id := r.Context().Value("user_id").(string)
	record := r.Context().Value("recordEvent").(func(response string))

	var data struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		utils.RespondError(w, record, "decode_error")
		return
	}

	err = db.SetKV(user_id, data.Key, data.Value)
	if err != nil {
		utils.RespondError(w, record, "database_error")
		return
	}

	record("[Empty Response]")

	w.WriteHeader(http.StatusCreated)
}
