package auth

import (
	"context"
	"testing"

	database "github.com/polyfire/api/db"
	"github.com/polyfire/api/middlewares"
	"github.com/polyfire/api/utils"
)

func MockGetUserIDFromProjectAuthID(_ string, _ string) (*string, error) {
	userID := "this-is-a-test-user-id"

	return &userID, nil
}

func TestExchangeToken(t *testing.T) {
	utils.SetLogLevel("WARN")

	MockGetUserFromToken := func(_ context.Context, token string, projectID string) (string, string, error) {
		if token != "this-is-a-test-token" {
			t.Fatalf(`ExchangeToken expected to pass token "this-is-a-test-token" but got "%s"`, token)
		}

		if projectID != "this-is-a-test-project-id" {
			t.Fatalf(`ExchangeToken expected to pass project-id "this-is-a-test-project-id" but got "%s"`, projectID)
		}
		return "test-user-id", "email@example.com", nil
	}

	ctx := context.WithValue(
		context.Background(),
		utils.ContextKeyDB,
		database.MockDatabase{
			MockGetUserIDFromProjectAuthID: MockGetUserIDFromProjectAuthID,
		},
	)

	project := database.Project{
		ID:     "this-is-a-test-project-id",
		Name:   "this-is-a-test-project-name",
		AuthID: "this-is-a-test-auth-id",
	}

	token, err := ExchangeToken(ctx, "this-is-a-test-token", project, MockGetUserFromToken)
	if err != nil {
		t.Fatalf(`ExchangeToken returned an error %v`, err)
	}

	claims, err := middlewares.ParseJWT(token)
	if err != nil {
		t.Fatalf(`ParseJWT error: %v`, err)
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		t.Fatalf(`Missing user_id in generated token`)
	}

	if userID != "this-is-a-test-user-id" {
		t.Fatalf(`Unexpected userID "%s"`, userID)
	}
}
