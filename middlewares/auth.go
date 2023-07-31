package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"os"

	jwt "github.com/golang-jwt/jwt/v5"
	router "github.com/julienschmidt/httprouter"
)

func AuthenticateRequest(r *http.Request) (context.Context, error) {
	if len(r.Header["X-Access-Token"]) == 0 {
		return nil, fmt.Errorf("403 forbidden")
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
		return nil, fmt.Errorf("403 forbidden")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid && err == nil {
		userId = claims["user_id"].(string)
	} else {
		return nil, fmt.Errorf("403 forbidden")
	}

	ctx := context.WithValue(r.Context(), "user_id", userId)

	return ctx, nil
}

func Auth(handler func(http.ResponseWriter, *http.Request, router.Params)) func(http.ResponseWriter, *http.Request, router.Params) {
	return func(w http.ResponseWriter, r *http.Request, params router.Params) {
		ctx, err := AuthenticateRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		handler(w, r.WithContext(ctx), params)
	}
}

func StreamAuth(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, err := AuthenticateRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}
