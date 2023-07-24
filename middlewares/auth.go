package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"os"

	jwt "github.com/golang-jwt/jwt/v5"
)

func Auth(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		if len(r.Header["X-Access-Token"]) == 0 {
			http.Error(w, "403 forbidden", http.StatusForbidden)
			return
		}
		access_token := r.Header["X-Access-Token"][0]
		token, err := jwt.Parse(access_token, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		var userId string

		if token == nil {
			http.Error(w, "403 forbidden", http.StatusForbidden)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid && err == nil {
			userId = claims["user_id"].(string)
		} else {
			http.Error(w, "403 forbidden", http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", userId)

		handler(w, r.WithContext(ctx))
	}
}
