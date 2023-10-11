package auth

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"
	router "github.com/julienschmidt/httprouter"
	db "github.com/polyfact/api/db"
	posthog "github.com/polyfact/api/posthog"
	"github.com/polyfact/api/utils"
)

func GetUserIdFromProjectAuthId(project string, auth_id string, email string) (*string, error) {
	client, err := db.CreateClient()
	if err != nil {
		return nil, err
	}

	var results []db.ProjectUser

	_, err = client.From("project_users").
		Select("*", "exact", false).
		Eq("auth_id", auth_id).
		Eq("project_id", project).
		ExecuteTo(&results)

	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}

	posthog.IdentifyUser(auth_id, results[0].ID, email)

	return &results[0].ID, nil
}

func CreateProjectUser(
	auth_id string,
	email string,
	project_id string,
	monthly_token_rate_limit *int,
	monthly_credit_rate_limit *int,
) (*string, error) {
	client, err := db.CreateClient()
	if err != nil {
		return nil, err
	}

	var result *db.ProjectUser

	_, err = client.From("project_users").Insert(db.ProjectUserInsert{
		AuthID:                 auth_id,
		ProjectID:              project_id,
		MonthlyTokenRateLimit:  monthly_token_rate_limit,
		MonthlyCreditRateLimit: monthly_credit_rate_limit,
	}, false, "", "", "exact").Single().ExecuteTo(&result)

	if err != nil {
		return nil, err
	}

	posthog.IdentifyUser(auth_id, result.ID, email)

	return &result.ID, nil
}

func ExchangeToken(
	token string,
	project db.Project,
	getUserFromToken func(token string, project_id string) (string, string, error),
) (string, error) {
	auth_id, email, err := getUserFromToken(token, project.ID)
	if err != nil {
		return "", err
	}

	user_id, err := GetUserIdFromProjectAuthId(project.ID, auth_id, email)
	if err != nil {
		return "", err
	}

	if user_id == nil {

		if !project.FreeUserInit {
			return "", fmt.Errorf("free_user_init_disabled")
		}
		user_id, err = CreateProjectUser(
			auth_id,
			email,
			project.ID,
			project.DefaultMonthlyTokenRateLimit,
			project.DefaultMonthlyCreditRateLimit,
		)
		if err != nil {
			return "", err
		}
	}

	unsigned_user_token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": *user_id,
	})

	user_token, err := unsigned_user_token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}

	return user_token, nil
}

func TokenExchangeHandler(
	getUserFromToken func(token string, project_id string) (string, string, error),
) func(w http.ResponseWriter, r *http.Request, ps router.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps router.Params) {
		record := r.Context().Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)
		project_id := ps.ByName("id")
		project, err := db.GetProjectByID(project_id)
		if err != nil || project == nil {
			utils.RespondError(w, record, "project_retrieval_error")
			return
		}

		if len(r.Header["Authorization"]) == 0 {
			utils.RespondError(w, record, "missing_authorization")
			return
		}

		auth_header := strings.Split(r.Header["Authorization"][0], " ")
		if len(auth_header) != 2 {
			utils.RespondError(w, record, "invalid_authorization_format")
			return
		}

		token := auth_header[1]

		user_token, err := ExchangeToken(token, *project, getUserFromToken)
		if err != nil {
			utils.RespondError(w, record, "token_exchange_failed")
			return
		}

		_, _ = w.Write([]byte(user_token))
	}
}
