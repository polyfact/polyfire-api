package middlewares

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"

	db "github.com/polyfire/api/db"
	posthog "github.com/polyfire/api/posthog"
	"github.com/polyfire/api/utils"
)

var isDevelopment = os.Getenv("APP_MODE") == "development"

func RecoverFromPanic(w http.ResponseWriter, r *http.Request) {
	record := r.Context().Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)
	if rec := recover(); rec != nil {
		errorMessage := getErrorMessage(rec)

		utils.RespondError(w, record, "unknown_error", errorMessage)
	}
}

func getErrorMessage(rec interface{}) string {
	if isDevelopment {
		switch v := rec.(type) {
		case error:
			return v.Error()
		case string:
			return v
		}
	}

	// For prod or for unhandled types
	return "Internal Server Error"
}

func AddRecord(r *http.Request) {
	var recordEventRequest utils.RecordRequestFunc = func(request string, response string, userID string, props ...utils.KeyValue) {
		go func() {
			pId, _ := db.GetProjectForUserId(userID)
			projectId := "00000000-0000-0000-0000-000000000000"

			if pId != nil {
				projectId = *pId
			}
			properties := make(map[string]string)
			properties["path"] = string(r.URL.Path)
			properties["projectId"] = projectId
			properties["requestBody"] = request
			properties["responseBody"] = response
			var error bool = false
			for _, element := range props {
				if element.Key == "Error" {
					error = true
				}
				properties[element.Key] = element.Value
			}
			posthog.Event("API Request", userID, properties)
			db.LogEvents(string(r.URL.Path), userID, projectId, request, response, error)
		}()
	}

	buf, _ := io.ReadAll(r.Body)
	rdr1 := io.NopCloser(bytes.NewBuffer(buf))

	r.Body = rdr1

	var recordEventWithUserID utils.RecordWithUserIDFunc = func(response string, userID string, props ...utils.KeyValue) {
		recordEventRequest(string(buf), response, userID, props...)
	}

	var recordEvent utils.RecordFunc = func(response string, props ...utils.KeyValue) {
		recordEventWithUserID(response, "", props...)
	}

	newCtx := context.WithValue(r.Context(), utils.ContextKeyRecordEvent, recordEvent)
	newCtx = context.WithValue(newCtx, utils.ContextKeyRecordEventRequest, recordEventRequest)
	newCtx = context.WithValue(newCtx, utils.ContextKeyRecordEventWithUserID, recordEventWithUserID)

	*r = *r.WithContext(newCtx)
}
