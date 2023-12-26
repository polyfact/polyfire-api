package kv

import (
	"encoding/json"
	"net/http"

	router "github.com/julienschmidt/httprouter"
	database "github.com/polyfire/api/db"
	"github.com/polyfire/api/utils"
)

func Get(w http.ResponseWriter, r *http.Request, _ router.Params) {
	db := r.Context().Value(utils.ContextKeyDB).(database.Database)
	userID := r.Context().Value(utils.ContextKeyUserID).(string)
	record := r.Context().Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)

	id := r.URL.Query().Get("key")

	if id == "" {
		utils.RespondError(w, record, "missing_id")
		return
	}

	store, _ := db.GetKV(userID, id)

	if store == nil || store.Value == "" {
		utils.RespondError(w, record, "data_not_found")
		return
	}

	record(store.Value)

	_, _ = w.Write([]byte(store.Value))
}

func Set(w http.ResponseWriter, r *http.Request, _ router.Params) {
	db := r.Context().Value(utils.ContextKeyDB).(database.Database)
	userID := r.Context().Value(utils.ContextKeyUserID).(string)
	record := r.Context().Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)

	var data struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		utils.RespondError(w, record, "decode_error")
		return
	}

	err = db.SetKV(userID, data.Key, data.Value)
	if err != nil {
		utils.RespondError(w, record, "database_error")
		return
	}

	record("[Empty Response]")

	w.WriteHeader(http.StatusCreated)
}

func Delete(w http.ResponseWriter, r *http.Request, _ router.Params) {
	db := r.Context().Value(utils.ContextKeyDB).(database.Database)
	userID := r.Context().Value(utils.ContextKeyUserID).(string)
	record := r.Context().Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)

	id := r.URL.Query().Get("key")

	if id == "" {
		utils.RespondError(w, record, "missing_id")
		return
	}

	err := db.DeleteKV(userID, id)
	if err != nil {
		utils.RespondError(w, record, "database_error")
		return
	}

	record("[Empty Response]")

	w.WriteHeader(http.StatusOK)
}

func List(w http.ResponseWriter, r *http.Request, _ router.Params) {
	db := r.Context().Value(utils.ContextKeyDB).(database.Database)
	userID := r.Context().Value(utils.ContextKeyUserID).(string)
	record := r.Context().Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)

	result, err := db.ListKV(userID)
	if err != nil {
		utils.RespondError(w, record, "database_error")
		return
	}

	response, _ := json.Marshal(result)
	record(string(response))

	_ = json.NewEncoder(w).Encode(result)
}
