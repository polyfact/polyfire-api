package auth

import (
	"context"
	"os"

	supa "github.com/nedpals/supabase-go"
)

func GetUserFromSupabaseToken(token string, _project_id string) (string, string, error) {
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

var ManagedSupabaseTokenExchangeHandler = TokenExchangeHandler(GetUserFromSupabaseToken)
