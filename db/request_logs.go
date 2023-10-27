package db

import (
	"database/sql"
	"strconv"
)

type Kind string

const (
	Completion Kind = "completion"
	Embed      Kind = "embedding"
	TTS        Kind = "tts"
)

type RequestLog struct {
	UserID           string `json:"user_id"`
	ModelName        string `json:"model_name"`
	InputTokenCount  *int   `json:"input_token_count"`
	OutputTokenCount *int   `json:"output_token_count"`
	Credits          int    `json:"credits"`
	Kind             Kind   `json:"kind"`
}

func tokenToCredit(provider_name string, model_name string, input_token_count int, output_token_count int) int {
	switch provider_name {
	case "openai":
		switch model_name {
		case "gpt-3.5-turbo":
			return (input_token_count * 15) + (output_token_count * 20)
		case "gpt-3.5-turbo-16k":
			return (input_token_count * 30) + (output_token_count * 40)
		case "gpt-4":
			return (input_token_count * 300) + (output_token_count * 600)
		case "gpt-4-32k":
			return (input_token_count * 600) + (output_token_count * 1200)
		case "text-embedding-ada-002":
			return input_token_count * 1
		case "dalle-2":
			return 200000
		}
	case "llama":
		return 0
	case "cohere":
		return (input_token_count + output_token_count) * 150
	}
	return 0
}

func LogRequests(
	eventID string,
	user_id string,
	provider_name string,
	model_name string,
	input_token_count int,
	output_token_count int,
	kind Kind,
	countCredits bool,
) {
	if kind == "" {
		kind = "completion"
	}
	var credits int

	if countCredits {
		credits = tokenToCredit(provider_name, model_name, input_token_count, output_token_count)
	} else {
		credits = 0
	}

	err := DB.Exec(
		"INSERT INTO request_logs (user_id, model_name, kind, input_token_count, output_token_count, credits, event_id) VALUES (?::uuid, ?, ?, ?, ?, ?, try_cast_uuid(?))",
		user_id,
		model_name,
		kind,
		input_token_count,
		output_token_count,
		credits,
		eventID,
	).Error
	if err != nil {
		panic(err)
	}
}

func LogRequestsCredits(
	eventID string,
	user_id string,
	provider_name string,
	model_name string,
	credits int,
	input_token_count int,
	output_token_count int,
	kind Kind,
) {
	err := DB.Exec(
		"INSERT INTO request_logs (user_id, model_name, kind, input_token_count, output_token_count, credits, event_id) VALUES (?::uuid, ?, ?, ?, ?, ?, try_cast_uuid(?))",
		user_id,
		model_name,
		kind,
		input_token_count,
		output_token_count,
		credits,
		eventID,
	).Error
	if err != nil {
		panic(err)
	}
}

type UsageParams struct {
	UserID string `json:"userid"`
}

func GetUserIdMonthlyTokenUsage(user_id string) (int, error) {
	params := UsageParams{
		UserID: user_id,
	}

	client, err := CreateClient()
	if err != nil {
		return 0, err
	}

	response := client.Rpc("get_monthly_token_usage", "", params)

	if response == "null" {
		return 0, nil
	}

	usage, err := strconv.Atoi(response)
	if err != nil {
		return 0, err
	}

	return usage, nil
}

func GetUserIdMonthlyCreditUsage(user_id string) (int, error) {
	params := UsageParams{
		UserID: user_id,
	}

	client, err := CreateClient()
	if err != nil {
		return 0, err
	}

	response := client.Rpc("get_monthly_credit_usage", "", params)

	if response == "null" {
		return 0, nil
	}

	usage, err := strconv.Atoi(response)
	if err != nil {
		return 0, err
	}

	return usage, nil
}

type Event struct {
	Path         string `json:"path"`
	Error        bool   `json:"error"`
	UserID       string `json:"user_id"`
	ProjectID    string `json:"project_id"`
	RequestBody  string `json:"request_body"`
	ResponseBody string `json:"response_body"`
	PromptID     string `json:"prompt_id"`
	Type         string `json:"type"`
}

func LogEvents(
	id string,
	path string,
	userId string,
	projectId string,
	requestBody string,
	responseBody string,
	error bool,
	promptID string,
	eventType string,
) {
	err := DB.Exec(
		`INSERT INTO events (id, path, user_id, project_id, request_body, response_body, error, prompt_id, type)
		VALUES (
			@id,
			@path,
			(CASE WHEN @user_id = '' THEN NULL ELSE @user_id END)::uuid,
			@project_id,
			@request_body,
			@response_body,
			@error,
			(SELECT id FROM prompts WHERE id = try_cast_uuid(@prompt_id) OR slug = @prompt_id)::uuid,
			@type
		)`,
		sql.Named("id", id),
		sql.Named("path", path),
		sql.Named("user_id", userId),
		sql.Named("project_id", projectId),
		sql.Named("request_body", requestBody),
		sql.Named("response_body", responseBody),
		sql.Named("error", error),
		sql.Named("prompt_id", promptID),
		sql.Named("type", eventType),
	).Error
	if err != nil {
		panic(err)
	}
}
