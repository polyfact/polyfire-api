package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"
	router "github.com/julienschmidt/httprouter"
	supa "github.com/nedpals/supabase-go"
	db "github.com/polyfact/api/db"
	posthog "github.com/polyfact/api/posthog"
	"github.com/polyfact/api/utils"
)

func GetAuthIdFromToken(token string) (string, string, error) {
	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	supabase := supa.CreateClient(supabaseUrl, supabaseKey)

	ctx := context.Background()
	user, err := supabase.Auth.User(
		ctx,
		token,
	)
	if err != nil {
		return "", "", err
	}

	return user.ID, user.Email, nil
}

func GetUserIdFromTokenProject(token string, project string) (*string, error) {
	auth_id, email, err := GetAuthIdFromToken(token)
	if err != nil {
		return nil, err
	}

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

func CreateProjectUser(token string, project_id string, monthly_token_rate_limit *int) (*string, error) {
	auth_id, email, err := GetAuthIdFromToken(token)
	if err != nil {
		return nil, err
	}

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

func TokenExchangeHandler(w http.ResponseWriter, r *http.Request, ps router.Params) {
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

	project, err := db.GetProjectByID(project_id)
	if err != nil {
		utils.RespondError(w, record, "project_retrieval_error")
		return
	}

	token := auth_header[1]

	user_id, err := GetUserIdFromTokenProject(token, project_id)
	if err != nil {
		utils.RespondError(w, record, "token_exchange_failed")
		return
	}

	if user_id == nil {
		if project.FreeUserInit == false {
			utils.RespondError(w, record, "free_user_init_disabled")
			return
		}
		user_id, err = CreateProjectUser(token, project_id, project.DefaultMonthlyTokenRateLimit)
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

type UserRateLimitResponse struct {
	Usage     int  `json:"usage"`
	RateLimit *int `json:"rate_limit"`
}

func UserRateLimit(w http.ResponseWriter, r *http.Request, _ router.Params) {
	user_id := r.Context().Value("user_id").(string)
	record := r.Context().Value("recordEvent").(utils.RecordFunc)

	tokenUsage, err := db.GetUserIdMonthlyTokenUsage(user_id)
	if err != nil {
		utils.RespondError(w, record, "internal_error")
		return
	}

	projectUser, err := db.GetProjectUserByID(user_id)
	var rateLimit *int = nil
	if projectUser != nil && err == nil {
		rateLimit = projectUser.MonthlyTokenRateLimit
	}

	result := UserRateLimitResponse{
		Usage:     tokenUsage,
		RateLimit: rateLimit,
	}

	response, _ := json.Marshal(&result)
	record(string(response))

	json.NewEncoder(w).Encode(result)
}
