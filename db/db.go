package db

import (
	"os"

	supa "github.com/nedpals/supabase-go"
)

type RequestLog struct {
	UserID           string `json:"user_id"`
	ModelName        string `json:"model_name"`
	InputTokenCount  *int   `json:"input_token_count"`
	OutputTokenCount *int   `json:"output_token_count"`
}

func toRef[T interface{}](n T) *T {
	return &n
}

func LogRequests(user_id string, model_name string, input_token_count int, output_token_count int) {
	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	supabase := supa.CreateClient(supabaseUrl, supabaseKey)

	row := RequestLog{
		UserID:           user_id,
		ModelName:        model_name,
		InputTokenCount:  toRef(input_token_count),
		OutputTokenCount: toRef(output_token_count),
	}

	var results []RequestLog
	err := supabase.DB.From("request_logs").Insert(row).Execute(&results)

	if err != nil {
		panic(err)
	}
}
