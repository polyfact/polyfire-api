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

func toRef[T interface{}](n T) *T {
	return &n
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
) {
	supabase, err := CreateClient()
	if err != nil {
		panic(err)
	}

	if kind == "" {
		kind = "completion"
	}

	row := RequestLog{
		UserID:           user_id,
		ModelName:        model_name,
		Kind:             kind,
		InputTokenCount:  toRef(input_token_count),
		OutputTokenCount: toRef(output_token_count),
		Credits:          tokenToCredit(provider_name, model_name, input_token_count, output_token_count),
	}

	_, _, err = supabase.From("request_logs").Insert(row, false, "", "", "exact").Execute()
	if err != nil {
		panic(err)
	}
}

func LogRequestsCredits(
	user_id string,
	provider_name string,
	model_name string,
	credits int,
	kind Kind,
) {
	supabase, err := CreateClient()
	if err != nil {
		panic(err)
	}

	row := RequestLog{
		UserID:    user_id,
		ModelName: model_name,
		Kind:      kind,
		Credits:   credits,
	}

	_, _, err = supabase.From("request_logs").Insert(row, false, "", "", "exact").Execute()
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
	supabase, err := CreateClient()
	if err != nil {
		panic(err)
	}

	row := Event{
		Path:         path,
		UserID:       userId,
		ProjectID:    projectId,
		RequestBody:  requestBody,
		ResponseBody: responseBody,
	}

	_, _, err = supabase.From("events").Insert(row, false, "", "", "exact").Execute()
	if err != nil {
		panic(err)
	}
}
