package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/mail"
	"strings"

	database "github.com/polyfire/api/db"
	"github.com/polyfire/api/utils"
)

func getUserFromAnonymousToken(ctx context.Context, token string, projectID string) (string, string, error) {
	db := ctx.Value(utils.ContextKeyDB).(database.DB)
	project, err := db.GetProjectByID(projectID)
	if err != nil {
		return "", "", err
	}

	if !project.AllowAnonymousAuth {
		return "", "", fmt.Errorf("Anonymous auth is not allowed for this project")
	}

	var email string
	if token == "auto" {
		email, err = db.GetDevEmail(projectID)
		if err != nil {
			return "", "", err
		}
	} else {

		emailBytes, err := base64.URLEncoding.DecodeString(token)
		if err != nil {
			return "", "", err
		}

		email = strings.TrimSpace(string(emailBytes))
	}

	reEncoded := base64.URLEncoding.EncodeToString(
		[]byte(email),
	) // This needs to be re-encoded since it has been trimmed by the previous line

	_, err = mail.ParseAddress(email)
	if err != nil {
		return "", "", err
	}

	return reEncoded + "@" + projectID, email, nil
}

var AnonymousTokenExchangeHandler = TokenExchangeHandler(getUserFromAnonymousToken)
