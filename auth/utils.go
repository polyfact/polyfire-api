package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"
	router "github.com/julienschmidt/httprouter"
	database "github.com/polyfire/api/db"
	posthog "github.com/polyfire/api/posthog"
	"github.com/polyfire/api/utils"
)

func ExchangeToken(
	ctx context.Context,
	token string,
	project database.Project,
	getUserFromToken func(ctx context.Context, token string, projectID string) (string, string, error),
) (string, error) {
	db := ctx.Value(utils.ContextKeyDB).(database.Database)
	authID, email, err := getUserFromToken(ctx, token, project.ID)
	if err != nil {
		return "", err
	}

	userID, err := db.GetUserIDFromProjectAuthID(project.ID, authID)
	if err != nil {
		return "", err
	}

	if userID == nil {
		if !project.FreeUserInit {
			return "", fmt.Errorf("free_user_init_disabled")
		}
		log.Println("[INFO] Creating user on project", project.ID)
		userID, err = db.CreateProjectUser(
			authID,
			project.ID,
			project.DefaultMonthlyCreditRateLimit,
		)
		if err != nil {
			return "", err
		}

	}

	posthog.IdentifyUser(authID, *userID, email)

	unsignedUserToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": *userID,
	})

	userToken, err := unsignedUserToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}

	return userToken, nil
}

func TokenExchangeHandler(
	getUserFromToken func(ctx context.Context, token string, projectID string) (string, string, error),
) func(w http.ResponseWriter, r *http.Request, ps router.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps router.Params) {
		db := r.Context().Value(utils.ContextKeyDB).(database.Database)
		record := r.Context().Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)
		projectID := ps.ByName("id")
		project, err := db.GetProjectByID(projectID)
		if err != nil || project == nil {
			utils.RespondError(w, record, "project_retrieval_error")
			return
		}

		if len(r.Header["Authorization"]) == 0 {
			utils.RespondError(w, record, "missing_authorization")
			return
		}

		authHeader := strings.Split(r.Header["Authorization"][0], " ")
		if len(authHeader) != 2 {
			utils.RespondError(w, record, "invalid_authorization_format")
			return
		}

		token := authHeader[1]

		userToken, err := ExchangeToken(r.Context(), token, *project, getUserFromToken)
		if err != nil {
			log.Println("[ERROR]", err)
			utils.RespondError(w, record, "token_exchange_failed")
			return
		}

		_, _ = w.Write([]byte(userToken))
	}
}
