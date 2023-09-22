package db

import (
	"time"
)

type PromptLikeInput struct {
	UserId   string `json:"user_id"`
	PromptId string `json:"prompt_id"`
}

type PromptLikeOutput struct {
	UserId   string `json:"user_id"`
	PromptId string `json:"prompt_id"`
	Like     bool   `json:"like"`
}

type PromptLike struct {
	UserId    string    `json:"user_id"`
	PromptId  string    `json:"prompt_id"`
	UpdatedAt time.Time `json:"updated_at"`
}

func GetPromptLikeByUserId(input PromptLikeInput) ([]PromptLike, error) {
	client, err := CreateClient()

	if err != nil {
		return nil, err
	}

	var result []PromptLike
	_, err = client.From("prompts_likes").Select("*", "exact", false).Eq("prompt_id", input.PromptId).Eq("user_id", input.UserId).Limit(1, "").ExecuteTo(&result)

	return result, err
}

func AddPromptLike(input PromptLikeInput) ([]PromptLike, error) {
	client, err := CreateClient()

	if err != nil {
		return nil, err
	}

	var result []PromptLike
	_, err = client.From("prompts_likes").Insert(input, false, "", "", "exact").ExecuteTo(&result)

	return result, err
}

func RemovePromptLike(input PromptLikeInput) ([]PromptLike, error) {
	client, err := CreateClient()
	if err != nil {
		return nil, err
	}

	var result []PromptLike
	_, err = client.From("prompts_likes").Delete("", "").Eq("user_id", input.UserId).Eq("prompt_id", input.PromptId).ExecuteTo(&result)

	return result, err
}
