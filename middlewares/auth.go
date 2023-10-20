package middlewares

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	jwt "github.com/golang-jwt/jwt/v5"
	router "github.com/julienschmidt/httprouter"
	"github.com/polyfire/api/db"
	"github.com/polyfire/api/utils"
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

func createUserContext(r *http.Request, userID string, user *db.UserInfos, rateLimitStatus db.RateLimitStatus) context.Context {
	recordEventWithUserID := r.Context().Value(utils.ContextKeyRecordEventWithUserID).(utils.RecordWithUserIDFunc)
	newCtx := context.WithValue(r.Context(), utils.ContextKeyUserID, userID)
	newCtx = context.WithValue(newCtx, utils.ContextKeyRateLimitStatus, rateLimitStatus)
	if user != nil {
		newCtx = context.WithValue(newCtx, utils.ContextKeyProjectID, user.ProjectId)
		if user.OpenaiToken != "" {
			newCtx = context.WithValue(newCtx, utils.ContextKeyOpenAIToken, user.OpenaiToken)
			if user.OpenaiOrg != "" {
				newCtx = context.WithValue(newCtx, utils.ContextKeyOpenAIOrg, user.OpenaiOrg)
			}
		}
		if user.ReplicateToken != "" {
			newCtx = context.WithValue(newCtx, utils.ContextKeyReplicateToken, user.ReplicateToken)
		}
	}

	var recordEvent utils.RecordFunc = func(response string, props ...utils.KeyValue) {
		recordEventWithUserID(response, userID, props...)
	}

	newCtx = context.WithValue(newCtx, utils.ContextKeyRecordEvent, recordEvent)

	return newCtx
}

func authenticateAndHandle(
	w http.ResponseWriter,
	r *http.Request,
	params router.Params,
	token string,
	handler func(http.ResponseWriter, *http.Request, router.Params),
) {
	record := r.Context().Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)
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
	if ok {
		version = int(versionJSON)
	}

	log.Println("DB Version / Rate Limit Status In Context")
	user, rateLimitStatus, err := db.CheckDBVersionRateLimit(userID, version)

	if err == db.DBVersionMismatch {
		utils.RespondError(w, record, "invalid_token")
		return
	}

	if err != nil {
		utils.RespondError(w, record, "database_error")
		return
	}

	if len(user.AuthorizedDomains) != 0 {
		originHeader := r.Header.Get("Origin")

		u, err := url.Parse(originHeader)
		if err != nil {
			utils.RespondError(w, record, "invalid_origin")
			return
		}
		origin := u.Hostname()
		if u.Port() != "" {
			origin = origin + ":" + u.Port()
		}

		if !utils.ContainsString(user.AuthorizedDomains, origin) {
			utils.RespondError(w, record, "invalid_origin")
			return
		}
	}

	ctx := createUserContext(r, userID, user, rateLimitStatus)
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
