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

func CreateProjectUser(auth_id string, email string, project_id string, monthly_token_rate_limit *int) (*string, error) {
	client, err := db.CreateClient()
	if err != nil {
		return nil, err
	}

	var result *db.ProjectUser

	_, err = client.From("project_users").Insert(db.ProjectUserInsert{
		AuthID:                auth_id,
		ProjectID:             project_id,
		MonthlyTokenRateLimit: monthly_token_rate_limit,
	}, false, "", "", "exact").Single().ExecuteTo(&result)

	if err != nil {
		return nil, err
	}

	posthog.IdentifyUser(auth_id, result.ID, email)

	return &result.ID, nil
}

func TokenExchangeHandler(
	getUserFromToken func(token string, project_id string) (string, string, error),
) func(w http.ResponseWriter, r *http.Request, ps router.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps router.Params) {
		record := r.Context().Value("recordEvent").(utils.RecordFunc)
		project_id := ps.ByName("id")

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

		auth_id, email, err := getUserFromToken(token, project_id)
		if err != nil {
			fmt.Println(err)
			utils.RespondError(w, record, "token_exchange_failed")
			return
		}

		user_id, err := GetUserIdFromProjectAuthId(project_id, auth_id, email)
		if err != nil {
			fmt.Println(err)
			utils.RespondError(w, record, "token_exchange_failed")
			return
		}

		if user_id == nil {
			project, err := db.GetProjectByID(project_id)
			if err != nil {
				utils.RespondError(w, record, "project_retrieval_error")
				return
			}
			if project.FreeUserInit == false {
				utils.RespondError(w, record, "free_user_init_disabled")
				return
			}
			user_id, err = CreateProjectUser(auth_id, email, project_id, project.DefaultMonthlyTokenRateLimit)
			if err != nil {
				utils.RespondError(w, record, "project_user_creation_failed")
				return
			}
		}

		unsigned_user_token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": *user_id,
		})

		user_token, err := unsigned_user_token.SignedString([]byte(os.Getenv("JWT_SECRET")))
		if err != nil {
			utils.RespondError(w, record, "token_signature_error")
			return
		}

		w.Write([]byte(user_token))
	}
}
