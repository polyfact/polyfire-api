package auth

import (
	"encoding/base64"
	"fmt"
	"net/mail"
	"strings"

	db "github.com/polyfire/api/db"
)

func getUserFromAnonymousToken(token string, project_id string) (string, string, error) {
	project, err := db.GetProjectByID(project_id)
	if err != nil {
		return "", "", err
	}

	if !project.AllowAnonymousAuth {
		return "", "", fmt.Errorf("Anonymous auth is not allowed for this project")
	}

	var email string
	if token == "auto" {
		email, err = db.GetDevEmail(project_id)
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

	re_encoded := base64.URLEncoding.EncodeToString(
		[]byte(email),
	) // This needs to be re-encoded since it has been trimmed by the previous line

	_, err = mail.ParseAddress(email)
	if err != nil {
		return "", "", err
	}

	return re_encoded + "@" + project_id, email, nil
}

var AnonymousTokenExchangeHandler = TokenExchangeHandler(getUserFromAnonymousToken)
