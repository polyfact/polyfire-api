package middlewares

import (
	"net/http"
	"os"

	"github.com/polyfact/api/utils"
)

var isDevelopment = os.Getenv("APP_MODE") == "development"

func RecoverFromPanic(w http.ResponseWriter) {
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
