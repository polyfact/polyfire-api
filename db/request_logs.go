package db

import (
	"database/sql"
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

func tokenToCredit(providerName string, modelName string, inputTokenCount int, outputTokenCount int) int {
	switch providerName {
	case "openai":
		switch modelName {
		case "gpt-3.5-turbo":
			return (inputTokenCount * 15) + (outputTokenCount * 20)
		case "gpt-3.5-turbo-16k":
			return (inputTokenCount * 30) + (outputTokenCount * 40)
		case "gpt-4":
			return (inputTokenCount * 300) + (outputTokenCount * 600)
		case "gpt-4-32k":
			return (inputTokenCount * 600) + (outputTokenCount * 1200)
		case "text-embedding-ada-002":
			return inputTokenCount * 1
		case "dalle-2":
			return 200000
		}
	case "llama":
		return 0
	case "cohere":
		return (inputTokenCount + outputTokenCount) * 150
	}
	return 0
}

func LogRequests(
	eventID string,
	userID string,
	providerName string,
	modelName string,
	inputTokenCount int,
	outputTokenCount int,
	kind Kind,
	countCredits bool,
) {
	if kind == "" {
		kind = "completion"
	}
	var credits int

	if countCredits {
		credits = tokenToCredit(providerName, modelName, inputTokenCount, outputTokenCount)
	} else {
		credits = 0
	}

	err1 := RemoveCreditsFromDev(userID, credits)

	err2 := DB.Exec(
		"INSERT INTO request_logs (user_id, model_name, kind, input_token_count, output_token_count, credits, event_id) VALUES (?::uuid, ?, ?, ?, ?, ?, try_cast_uuid(?))",
		userID,
		modelName,
		kind,
		inputTokenCount,
		outputTokenCount,
		credits,
		eventID,
	).Error
	if err1 != nil {
		panic(err1)
	}
	if err2 != nil {
		panic(err2)
	}
}

func LogRequestsCredits(
	eventID string,
	userID string,
	modelName string,
	credits int,
	inputTokenCount int,
	outputTokenCount int,
	kind Kind,
) {
	err1 := RemoveCreditsFromDev(userID, credits)
	err2 := DB.Exec(
		"INSERT INTO request_logs (user_id, model_name, kind, input_token_count, output_token_count, credits, event_id) VALUES (?::uuid, ?, ?, ?, ?, ?, try_cast_uuid(?))",
		userID,
		modelName,
		kind,
		inputTokenCount,
		outputTokenCount,
		credits,
		eventID,
	).Error
	if err1 != nil {
		panic(err1)
	}
	if err2 != nil {
		panic(err2)
	}
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
	userID string,
	projectID string,
	requestBody string,
	responseBody string,
	error bool,
	promptID string,
	eventType string,
	orginDomain string,
) {
	err := DB.Exec(
		`INSERT INTO events (id, path, user_id, project_id, request_body, response_body, error, prompt_id, type, origin_domain)
		VALUES (
			@id,
			@path,
			(CASE WHEN @user_id = '' THEN NULL ELSE @user_id END)::uuid,
			@project_id,
			@request_body,
			@response_body,
			@error,
			(SELECT id FROM prompts WHERE id = try_cast_uuid(@prompt_id) OR slug = @prompt_id LIMIT 1)::uuid,
			@type,
			@origin_domain
		)`,
		sql.Named("id", id),
		sql.Named("path", path),
		sql.Named("user_id", userID),
		sql.Named("project_id", projectID),
		sql.Named("request_body", requestBody),
		sql.Named("response_body", responseBody),
		sql.Named("error", error),
		sql.Named("prompt_id", promptID),
		sql.Named("type", eventType),
		sql.Named("origin_domain", orginDomain),
	).Error
	if err != nil {
		panic(err)
	}
}
