package auth

import (
	_ "embed"
	"encoding/json"
	"fmt"

	jwt "github.com/golang-jwt/jwt/v5"
	db "github.com/polyfire/api/db"
)

//go:embed firebase_public-keys.json
var public_keys []byte

func getUserFromFirebaseToken(firebase_token string, project_id string) (string, string, error) {
	project, err := db.GetProjectByID(project_id)
	if err != nil {
		return "", "", err
	}

	var objmap map[string]string
	err = json.Unmarshal(public_keys, &objmap)
	if err != nil {
		return "", "", err
	}

	token, err := jwt.Parse(firebase_token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		public_key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(objmap[token.Header["kid"].(string)]))
		if err != nil {
			return nil, err
		}
		return public_key, nil
	})
	if err != nil {
		return "", "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		firebase := claims["firebase"].(map[string]interface{})
		identities := firebase["identities"].(map[string]interface{})
		emails := identities["email"].([]interface{})
		email := emails[0].(string)
		user_id := claims["user_id"].(string)

		if project.FirebaseProjectID != claims["aud"] {
			return "", "", fmt.Errorf("ProjectID Mismatch")
		}

		return user_id, email, nil
	} else {
		return "", "", fmt.Errorf("Invalid token")
	}
}

var ExternalFirebaseTokenExchangeHandler = TokenExchangeHandler(getUserFromFirebaseToken)
