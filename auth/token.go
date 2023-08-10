package auth

import (
	"context"
	"fmt"
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

type ProjectUser struct {
	ID        string `json:"id"`
	AuthID    string `json:"auth_id"`
	ProjectID string `json:"project_id"`
}

type ProjectUserInsert struct {
	AuthID    string `json:"auth_id"`
	ProjectID string `json:"project_id"`
}

type Project struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	AuthID       string `json:"auth_id"`
	FreeUserInit bool   `json:"free_user_init"`
}

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

	var results []ProjectUser

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

func GetProjectByID(id string) (*Project, error) {
	client, err := db.CreateClient()
	if err != nil {
		return nil, err
	}

	var result Project

	_, err = client.From("projects").
		Select("*", "exact", false).
		Eq("id", id).
		Single().
		ExecuteTo(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func CreateProjectUser(token string, project_id string) (*string, error) {
	auth_id, email, err := GetAuthIdFromToken(token)
	if err != nil {
		return nil, err
	}

	client, err := db.CreateClient()
	if err != nil {
		return nil, err
	}

	var result *ProjectUser

	_, err = client.From("project_users").Insert(ProjectUserInsert{
		AuthID:    auth_id,
		ProjectID: project_id,
	}, false, "", "", "exact").Single().ExecuteTo(&result)

	if err != nil {
		return nil, err
	}

	posthog.IdentifyUser(auth_id, result.ID, email)

	return &result.ID, nil
}

func TokenExchangeHandler(w http.ResponseWriter, r *http.Request, ps router.Params) {
	project_id := ps.ByName("id")

	if len(r.Header["Authorization"]) == 0 {
		utils.RespondError(w, "missing_authorization")
		return
	}

	auth_header := strings.Split(r.Header["Authorization"][0], " ")
	if len(auth_header) != 2 {
		utils.RespondError(w, "invalid_authorization_format")
		return
	}

	token := auth_header[1]

	user_id, err := GetUserIdFromTokenProject(token, project_id)
	if err != nil {
		utils.RespondError(w, "token_exchange_failed")
		return
	}

	if user_id == nil {
		project, err := GetProjectByID(project_id)
		if err != nil {
			utils.RespondError(w, "project_retrieval_error")
			return
		}
		if project.FreeUserInit == false {
			utils.RespondError(w, "free_user_init_disabled")
			return
		}
		user_id, err = CreateProjectUser(token, project_id)
		if err != nil {
			utils.RespondError(w, "project_user_creation_failed")
			return
		}
	}

	unsigned_user_token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": *user_id,
	})

	user_token, err := unsigned_user_token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		fmt.Println(err)
		utils.RespondError(w, "token_signature_error")
		return
	}

	w.Write([]byte(user_token))
}
