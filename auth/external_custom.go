package auth

import (
	_ "embed"
	"fmt"

	jwt "github.com/golang-jwt/jwt/v5"
	db "github.com/polyfire/api/db"
)

func getUserFromCustomSignature(custom_token string, project_id string) (string, string, error) {
	project, err := db.GetProjectByID(project_id)
	if err != nil {
		return "", "", err
	}

	public_key_bytes := []byte(project.CustomAuthPublicKey)

	token, err := jwt.Parse(custom_token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		public_key, err := jwt.ParseRSAPublicKeyFromPEM(public_key_bytes)
		if err != nil {
			return nil, err
		}
		return public_key, nil
	})
	if err != nil {
		return "", "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		project_id := claims["iss"].(string)
		aud := claims["aud"].(string)
		user_id := claims["sub"].(string)

		if aud != "polyfire_custom_auth_api" {
			return "", "", fmt.Errorf("Invalid audience")
		}

		if project_id != project.ID {
			return "", "", fmt.Errorf("Invalid project id")
		}

		return user_id + "@" + project_id, user_id + "@" + project_id, nil
	} else {
		return "", "", fmt.Errorf("Invalid token")
	}
}

var ExternalCustomTokenExchangeHandler = TokenExchangeHandler(getUserFromCustomSignature)
