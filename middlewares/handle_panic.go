package middlewares

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/google/uuid"

	router "github.com/julienschmidt/httprouter"

	database "github.com/polyfire/api/db"
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

func AddRecord(r *http.Request, eventType utils.EventType) {
	db := r.Context().Value(utils.ContextKeyDB).(database.Database)
	eventID := uuid.New().String()

	originHeader := r.Header.Get("Origin")
	origin := ""

	if originHeader != "" {
		u, err := url.Parse(originHeader)
		if err == nil {
			origin = u.Hostname()
			if u.Port() != "" {
				origin = origin + ":" + u.Port()
			}
		}
	}

	var recordEventRequest utils.RecordRequestFunc = func(request string, response string, userID string, props ...utils.KeyValue) {
		go func() {
			pID, _ := db.GetProjectForUserID(userID)
			projectID := "00000000-0000-0000-0000-000000000000"

			if pID != nil {
				projectID = *pID
			}
			properties := make(map[string]string)
			properties["path"] = string(r.URL.Path)
			properties["projectID"] = projectID
			properties["requestBody"] = request
			properties["responseBody"] = response
			var isError bool = false
			var promptID string = ""
			for _, element := range props {
				if element.Key == "Error" {
					isError = true
				}
				if element.Key == "PromptID" {
					promptID = element.Value
				}
				properties[element.Key] = element.Value
			}
			posthog.Event("API Request", userID, properties)
			db.LogEvents(
				eventID,
				string(r.URL.Path),
				userID,
				projectID,
				request,
				response,
				isError,
				promptID,
				string(eventType),
				origin,
			)
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
	newCtx = context.WithValue(newCtx, utils.ContextKeyEventID, eventID)
	newCtx = context.WithValue(newCtx, utils.ContextKeyOriginDomain, origin)
	newCtx = context.WithValue(newCtx, utils.ContextKeyRecordEventRequest, recordEventRequest)
	newCtx = context.WithValue(newCtx, utils.ContextKeyRecordEventWithUserID, recordEventWithUserID)

	*r = *r.WithContext(newCtx)
}

func Record(
	eventType utils.EventType,
	handler func(http.ResponseWriter, *http.Request, router.Params),
) func(http.ResponseWriter, *http.Request, router.Params) {
	return func(w http.ResponseWriter, r *http.Request, params router.Params) {
		AddRecord(r, eventType)
		handler(w, r, params)
	}
}
