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
)

type ProjectUser struct {
	ID        string `json:"id"`
	AuthID    string `json:"auth_id"`
	ProjectID string `json:"project_id"`
}

func GetAuthIdFromToken(token string) (string, error) {
	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	supabase := supa.CreateClient(supabaseUrl, supabaseKey)

	ctx := context.Background()
	user, err := supabase.Auth.User(
		ctx,
		token,
	)
	if err != nil {
		return "", err
	}

	return user.ID, nil
}

func GetUserIdFromTokenProject(token string, project string) (string, error) {
	auth_id, err := GetAuthIdFromToken(token)
	if err != nil {
		return "", err
	}

	client, err := db.CreateClient()
	if err != nil {
		return "", err
	}

	var result *ProjectUser

	_, err = client.From("project_users").
		Select("*", "exact", false).
		Eq("auth_id", auth_id).
		Eq("project_id", project).
		Single().
		ExecuteTo(&result)
	if err != nil {
		return "", err
	}

	return result.ID, nil
}

func TokenExchangeHandler(w http.ResponseWriter, r *http.Request, ps router.Params) {
	project_id := ps.ByName("id")

	if len(r.Header["Authorization"]) == 0 {
		fmt.Println("ABCDEF1")
		http.Error(w, "403 forbidden", http.StatusForbidden)
		return
	}

	auth_header := strings.Split(r.Header["Authorization"][0], " ")
	if len(auth_header) != 2 {
		fmt.Println("ABCDEF2")
		http.Error(w, "403 forbidden", http.StatusForbidden)
		return
	}

	token := auth_header[1]

	user_id, err := GetUserIdFromTokenProject(token, project_id)
	if err != nil {
		fmt.Println("ABCDEF3")
		http.Error(w, "403 forbidden", http.StatusForbidden)
		return
	}

	unsigned_user_token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user_id,
	})

	user_token, err := unsigned_user_token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		fmt.Println(err)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(user_token))
}
