package auth

import (
	"fmt"

	jwt "github.com/golang-jwt/jwt/v5"
	db "github.com/polyfire/api/db"
)

func getUserFromCustomSignature(customToken string, projectID string) (string, string, error) {
	project, err := db.GetProjectByID(projectID)
	if err != nil {
		return "", "", err
	}

	publicKeyBytes := []byte(project.CustomAuthPublicKey)

	token, err := jwt.Parse(customToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
		if err != nil {
			return nil, err
		}
		return publicKey, nil
	})
	if err != nil {
		return "", "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		projectID := claims["iss"].(string)
		tokenProject, err := db.GetProjectByID(projectID)
		if err != nil {
			return "", "", err
		}

		aud := claims["aud"].(string)
		userID := claims["sub"].(string)

		if aud != "polyfire_custom_auth_api" {
			return "", "", fmt.Errorf("Invalid audience")
		}

		if tokenProject == nil || tokenProject.ID != project.ID {
			return "", "", fmt.Errorf("Invalid project id")
		}

		return userID + "@" + tokenProject.ID, userID + "@" + tokenProject.ID, nil
	}
	return "", "", fmt.Errorf("Invalid token")
}

var ExternalCustomTokenExchangeHandler = TokenExchangeHandler(getUserFromCustomSignature)
