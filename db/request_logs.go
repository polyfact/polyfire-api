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
	Kind             Kind   `json:"kind"`
}

func toRef[T interface{}](n T) *T {
	return &n
}

func LogRequests(
	user_id string,
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
