package middlewares

import (
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	"github.com/polyfact/api/utils"
)

var isDevelopment = os.Getenv("APP_MODE") == "development"

func HandlePanic(handler httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		defer recoverFromPanic(w)

		handler(w, r, params)
	}
}

func recoverFromPanic(w http.ResponseWriter) {
	if rec := recover(); rec != nil {
		errorMessage := getErrorMessage(rec)

		utils.RespondError(w, "unknown_error", errorMessage)
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
