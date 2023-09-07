package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"os"

	jwt "github.com/golang-jwt/jwt/v5"
	router "github.com/julienschmidt/httprouter"
	"github.com/polyfact/api/db"
	"github.com/polyfact/api/utils"
)

func parseJWT(token string) (jwt.MapClaims, error) {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return nil, fmt.Errorf("Invalid token")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || !parsedToken.Valid {
		return nil, fmt.Errorf("Invalid token claims")
	}

	return claims, nil
}

func createContextWithUserID(r *http.Request, userID string) context.Context {
	recordEventWithUserID := r.Context().Value("recordEventWithUserID").(utils.RecordWithUserIDFunc)
	newCtx := context.WithValue(r.Context(), "user_id", userID)

	var recordEvent utils.RecordFunc
	recordEvent = func(response string, props ...utils.KeyValue) {
		recordEventWithUserID(response, userID, props...)
	}

	newCtx = context.WithValue(newCtx, "recordEvent", recordEvent)

	return newCtx
}

func authenticateAndHandle(
	w http.ResponseWriter,
	r *http.Request,
	params router.Params,
	token string,
	handler func(http.ResponseWriter, *http.Request, router.Params),
) {
	record := r.Context().Value("recordEvent").(utils.RecordFunc)
	if token == "" {
		utils.RespondError(w, record, "no_token")
		return
	}

	claims, err := parseJWT(token)
	if err != nil {
		utils.RespondError(w, record, "invalid_token")
		return
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		utils.RespondError(w, record, "missing_user_id")
		return
	}

	var version int = 0
	versionJSON, ok := claims["version"].(float64)
	fmt.Println(versionJSON)
	if ok {
		version = int(versionJSON)
	}

	dbVersion, err := db.GetVersionForUser(userID)
	if err != nil {
		utils.RespondError(w, record, "database_error")
		return
	}

	fmt.Println(version, dbVersion)
	if version != dbVersion {
		utils.RespondError(w, record, "invalid_token")
		return
	}

	ctx := createContextWithUserID(r, userID)
	handler(w, r.WithContext(ctx), params)
}

func Auth(
	handler func(http.ResponseWriter, *http.Request, router.Params),
) func(http.ResponseWriter, *http.Request, router.Params) {
	return func(w http.ResponseWriter, r *http.Request, params router.Params) {
		token := r.Header.Get("X-Access-Token")
		authenticateAndHandle(w, r, params, token, handler)
	}
}

func AuthStream(
	handler func(http.ResponseWriter, *http.Request, router.Params),
) func(http.ResponseWriter, *http.Request, router.Params) {
	return func(w http.ResponseWriter, r *http.Request, params router.Params) {
		token := r.URL.Query().Get("token")
		authenticateAndHandle(w, r, params, token, handler)
	}
}
