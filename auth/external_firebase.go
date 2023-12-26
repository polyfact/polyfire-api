package auth

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	jwt "github.com/golang-jwt/jwt/v5"
	database "github.com/polyfire/api/db"
	"github.com/polyfire/api/utils"
)

//go:embed firebase_public-keys.json
var publicKeys []byte

func getUserFromFirebaseToken(ctx context.Context, firebaseToken string, projectID string) (string, string, error) {
	db := ctx.Value(utils.ContextKeyDB).(database.Database)
	project, err := db.GetProjectByID(projectID)
	if err != nil {
		return "", "", err
	}

	var objmap map[string]string
	err = json.Unmarshal(publicKeys, &objmap)
	if err != nil {
		return "", "", err
	}

	token, err := jwt.Parse(firebaseToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		publicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(objmap[token.Header["kid"].(string)]))
		if err != nil {
			return nil, err
		}
		return publicKey, nil
	})
	if err != nil {
		return "", "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		firebase := claims["firebase"].(map[string]interface{})
		identities := firebase["identities"].(map[string]interface{})
		emails := identities["email"].([]interface{})
		email := emails[0].(string)
		userID := claims["user_id"].(string)

		if project.FirebaseProjectID != claims["aud"] {
			return "", "", fmt.Errorf("ProjectID Mismatch")
		}

		return userID, email, nil
	}
	return "", "", fmt.Errorf("Invalid token")
}

var ExternalFirebaseTokenExchangeHandler = TokenExchangeHandler(getUserFromFirebaseToken)
