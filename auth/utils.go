package auth

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"
	router "github.com/julienschmidt/httprouter"
	db "github.com/polyfire/api/db"
	posthog "github.com/polyfire/api/posthog"
	"github.com/polyfire/api/utils"
)

func GetUserIDFromProjectAuthID(project string, authID string, email string) (*string, error) {
	client, err := db.CreateClient()
	if err != nil {
		return nil, err
	}

	var results []db.ProjectUser

	_, err = client.From("project_users").
		Select("*", "exact", false).
		Eq("auth_id", authID).
		Eq("project_id", project).
		ExecuteTo(&results)

	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}

	posthog.IdentifyUser(authID, results[0].ID, email)

	return &results[0].ID, nil
}

func CreateProjectUser(
	authID string,
	email string,
	projectID string,
	monthlyCreditRateLimit *int,
) (*string, error) {
	client, err := db.CreateClient()
	if err != nil {
		return nil, err
	}

	var result *db.ProjectUser

	_, err = client.From("project_users").Insert(db.ProjectUserInsert{
		AuthID:                 authID,
		ProjectID:              projectID,
		MonthlyCreditRateLimit: monthlyCreditRateLimit,
	}, false, "", "", "exact").Single().ExecuteTo(&result)

	if err != nil {
		return nil, err
	}

	posthog.IdentifyUser(authID, result.ID, email)

	return &result.ID, nil
}

func ExchangeToken(
	token string,
	project db.Project,
	getUserFromToken func(token string, projectID string) (string, string, error),
) (string, error) {
	authID, email, err := getUserFromToken(token, project.ID)
	if err != nil {
		return "", err
	}

	userID, err := GetUserIDFromProjectAuthID(project.ID, authID, email)
	if err != nil {
		return "", err
	}

	if userID == nil {

		if !project.FreeUserInit {
			return "", fmt.Errorf("free_user_init_disabled")
		}
		userID, err = CreateProjectUser(
			authID,
			email,
			project.ID,
			project.DefaultMonthlyCreditRateLimit,
		)
		if err != nil {
			return "", err
		}
	}

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
	getUserFromToken func(token string, projectID string) (string, string, error),
) func(w http.ResponseWriter, r *http.Request, ps router.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps router.Params) {
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

		userToken, err := ExchangeToken(token, *project, getUserFromToken)
		if err != nil {
			fmt.Println(err)
			utils.RespondError(w, record, "token_exchange_failed")
			return
		}

		_, _ = w.Write([]byte(userToken))
	}
}
