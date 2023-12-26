package auth

import (
	"context"
	"os"

	supa "github.com/nedpals/supabase-go"
)

func GetUserFromSupabaseToken(_ context.Context, token string, _ string) (string, string, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	supabase := supa.CreateClient(supabaseURL, supabaseKey)

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
