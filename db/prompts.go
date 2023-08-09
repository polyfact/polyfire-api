package db

import (
	"time"
)

type Prompt struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Prompt      string    `json:"prompt"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	Like        int64     `json:"like,omitempty"`
	Use         int64     `json:"use,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
}

type PromptInsert struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Prompt      string   `json:"prompt"`
	Tags        []string `json:"tags,omitempty"`
}

type PromptUpdate struct {
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	Name        string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	Prompt      string    `json:"prompt,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	Likes       int64     `json:"likes,omitempty"`
	Use         int64     `json:"use,omitempty"`
}

func GetPromptById(id string) (*Prompt, error) {
	client, err := CreateClient()
	if err != nil {
		return nil, err
	}

	var result *Prompt

	_, err = client.From("prompts").Select("*", "exact", false).Eq("id", id).Single().ExecuteTo(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func GetPromptByName(name string) (*Prompt, error) {
	client, err := CreateClient()
	if err != nil {
		return nil, err
	}

	var result *Prompt

	_, err = client.From("prompts").Select("*", "exact", false).Eq("name", name).Single().ExecuteTo(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func GetAllPrompts() ([]Prompt, error) {
	client, err := CreateClient()
	if err != nil {
		return nil, err
	}

	var results []Prompt

	_, err = client.From("prompts").Select("*", "exact", false).ExecuteTo(&results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func CreatePrompt(input PromptInsert) (*Prompt, error) {
	client, err := CreateClient()
	if err != nil {
		return nil, err
	}

	var result *Prompt

	_, err = client.From("prompts").Insert(input, false, "", "", "exact").Single().ExecuteTo(&result)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func UpdatePrompt(id string, input PromptUpdate) (*Prompt, error) {
	client, err := CreateClient()
	if err != nil {
		return nil, err
	}

	input.UpdatedAt = time.Now()

	var result *Prompt

	_, err = client.From("prompts").Update(input, "", "").Eq("id", id).Single().ExecuteTo(&result)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func DeletePrompt(id string) error {
	client, err := CreateClient()
	if err != nil {
		return err
	}

	_, _, err = client.From("prompts").Delete("", "").Eq("id", id).Execute()

	if err != nil {
		return err
	}

	return nil
}
