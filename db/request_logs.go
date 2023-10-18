package db

import (
	"strconv"
)

type Kind string

const (
	Completion Kind = "completion"
	Embed      Kind = "embedding"
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
		"INSERT INTO request_logs (user_id, model_name, kind, input_token_count, output_token_count, credits) VALUES (?::uuid, ?, ?, ?, ?, ?)",
		user_id,
		model_name,
		kind,
		input_token_count,
		output_token_count,
		credits,
	).Error
	if err != nil {
		panic(err)
	}
}

func LogRequestsCredits(
	user_id string,
	provider_name string,
	model_name string,
	credits int,
	input_token_count int,
	output_token_count int,
	kind Kind,
) {
	err := DB.Exec(
		"INSERT INTO request_logs (user_id, model_name, kind, input_token_count, output_token_count, credits) VALUES (?::uuid, ?, ?, ?, ?, ?)",
		user_id,
		model_name,
		kind,
		input_token_count,
		output_token_count,
		credits,
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
}

func LogEvents(
	path string,
	userId string,
	projectId string,
	requestBody string,
	responseBody string,
) {
	err := DB.Exec(
		"INSERT INTO events (path, user_id, project_id, request_body, response_body) VALUES (?, ?::uuid, ?, ?, ?)",
		path,
		userId,
		projectId,
		requestBody,
		responseBody,
	).Error
	if err != nil {
		panic(err)
	}
}
