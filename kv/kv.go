package kv

import (
	"encoding/json"
	"net/http"

	router "github.com/julienschmidt/httprouter"
	db "github.com/polyfact/api/db"
)

func Get(w http.ResponseWriter, r *http.Request, _ router.Params) {
	user_id := r.Context().Value("user_id").(string)

	id := r.URL.Query().Get("key")

	if id == "" {
		http.Error(w, "Missing id", http.StatusBadRequest)
		return
	}

	store, _ := db.GetKV(user_id, id)

	if store == nil || store.Value == "" {
		http.Error(w, "", http.StatusNotFound)
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = db.SetKV(user_id, data.Key, data.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
