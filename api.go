package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"

	jwt "github.com/golang-jwt/jwt/v5"

	db "github.com/polyfact/api/db"
	llm "github.com/polyfact/api/llm"
	stt "github.com/polyfact/api/stt"
)

type GenerateRequestBody struct {
	Task     string `json:"task"`
	Provider string `json:"provider"`
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

	if input.Provider == "" {
		input.Provider = "openai"
	}

	provider, err := llm.NewLLMProvider(input.Provider)
	if err == llm.ErrUnknownModel {
		http.Error(w, "400 Unknown model provider", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	var result llm.Result
	result, err = provider.Generate(input.Task, &callback)

	w.Header()["Content-Type"] = []string{"application/json"}

	if err != nil {
		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

func transcribe(w http.ResponseWriter, r *http.Request) {
	_, p, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	boundary := p["boundary"]
	reader := multipart.NewReader(r.Body, boundary)
	part, err := reader.NextPart()
	if err == io.EOF {
		http.Error(w, "400 Bad Request", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
		return
	}
	file_buf_reader := bufio.NewReader(part)

	// The format doesn't seem to really matter
	res, err := stt.Transcribe(file_buf_reader, "mp3")
	if err != nil {
		log.Printf("%w", err)
		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(res)
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
	http.HandleFunc("/transcribe", authMiddleware(transcribe))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
