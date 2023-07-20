package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	jwt "github.com/golang-jwt/jwt/v5"

	db "github.com/polyfact/api/db"
	llm "github.com/polyfact/api/llm"
	memory "github.com/polyfact/api/memory"
)

type GenerateRequestBody struct {
	Task string `json:"task"`
}

func generate(w http.ResponseWriter, r *http.Request) {
	user_id := r.Context().Value("user_id").(string)

	if r.Method != "POST" {
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if len(r.Header["Content-Type"]) == 0 || r.Header["Content-Type"][0] != "application/json" {
		http.Error(w, "400 bad request", http.StatusBadRequest)
		return
	}

	var input GenerateRequestBody

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "400 bad request", http.StatusBadRequest)
		return
	}

	callback := func(model_name string, input_count int, output_count int) {
		db.LogRequests(user_id, model_name, input_count, output_count)
	}

	var result llm.Result
	result, err = llm.Generate(input.Task, &callback)

	w.Header()["Content-Type"] = []string{"application/json"}

	if err != nil {
		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

func authMiddleware(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
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

		var user_id string

		if token == nil {
			http.Error(w, "403 forbidden", http.StatusForbidden)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid && err == nil {
			user_id = claims["user_id"].(string)
		} else {
			http.Error(w, "403 forbidden", http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", user_id)

		handler(w, r.WithContext(ctx))
	}
}

func main() {
	log.Print("Starting the server on :8080")
	http.HandleFunc("/generate", authMiddleware(generate))
	http.HandleFunc("/memory", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			memory.Create(w, r)
		} else if r.Method == http.MethodPut {
			memory.Add(w, r)
		} else {
			http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	// http.HandleFunc("/memories", memory.Get)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
