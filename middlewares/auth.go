package middlewares

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	jwt "github.com/golang-jwt/jwt/v5"
	router "github.com/julienschmidt/httprouter"
	database "github.com/polyfire/api/db"
	"github.com/polyfire/api/utils"
)

func ParseJWT(token string) (jwt.MapClaims, error) {
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

func createUserContext(
	r *http.Request,
	userID string,
	user *database.UserInfos,
	rateLimitStatus database.RateLimitStatus,
	creditsStatus database.CreditsStatus,
) context.Context {
	recordEventWithUserID := r.Context().Value(utils.ContextKeyRecordEventWithUserID).(utils.RecordWithUserIDFunc)
	newCtx := context.WithValue(r.Context(), utils.ContextKeyUserID, userID)
	newCtx = context.WithValue(newCtx, utils.ContextKeyRateLimitStatus, rateLimitStatus)
	newCtx = context.WithValue(newCtx, utils.ContextKeyCreditsStatus, creditsStatus)
	if user != nil {
		newCtx = context.WithValue(newCtx, utils.ContextKeyProjectUserUsage, user.ProjectUserUsage)
		newCtx = context.WithValue(newCtx, utils.ContextKeyProjectUserRateLimit, user.ProjectUserRateLimit)
		newCtx = context.WithValue(newCtx, utils.ContextKeyProjectID, user.ProjectID)
		if user.OpenaiToken != "" {
			newCtx = context.WithValue(newCtx, utils.ContextKeyOpenAIToken, user.OpenaiToken)
			if user.OpenaiOrg != "" {
				newCtx = context.WithValue(newCtx, utils.ContextKeyOpenAIOrg, user.OpenaiOrg)
			}
		}
		if user.ReplicateToken != "" {
			newCtx = context.WithValue(newCtx, utils.ContextKeyReplicateToken, user.ReplicateToken)
		}
		if user.ElevenlabsToken != "" {
			newCtx = context.WithValue(newCtx, utils.ContextKeyElevenlabsToken, user.ElevenlabsToken)
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
	db := r.Context().Value(utils.ContextKeyDB).(database.Database)
	record := r.Context().Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)
	if token == "" {
		utils.RespondError(w, record, "no_token")
		return
	}

	claims, err := ParseJWT(token)
	if err != nil {
		utils.RespondError(w, record, "invalid_token")
		return
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		utils.RespondError(w, record, "missing_user_id")
		return
	}

	version := 0
	versionJSON, ok := claims["version"].(float64)
	if ok {
		version = int(versionJSON)
	}

	log.Println("DB Version / Rate Limit Status In Context")
	user, rateLimitStatus, creditsStatus, err := db.CheckDBVersionRateLimit(userID, version)

	if err == database.ErrDBVersionMismatch {
		utils.RespondError(w, record, "invalid_token")
		return
	}

	if err == database.ErrDevNotPremium {
		utils.RespondError(w, record, "dev_not_premium")
		return
	}

	if err != nil {
		fmt.Println(err)
		utils.RespondError(w, record, "database_error")
		return
	}

	if len(user.AuthorizedDomains) != 0 {
		originDomain := r.Context().Value(utils.ContextKeyOriginDomain).(string)
		if !utils.ContainsString(user.AuthorizedDomains, originDomain) && originDomain != "beta.polyfire.com" {
			utils.RespondError(w, record, "invalid_origin")
			return
		}
	}

	ctx := createUserContext(r, userID, user, rateLimitStatus, creditsStatus)
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
