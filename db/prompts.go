package db

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type StringArray []string

func (o *StringArray) Scan(src any) error {
	str, ok := src.(string)
	if !ok {
		return errors.New("src value cannot cast to string")
	}
	*o = strings.Split(strings.Trim(str, "{}"), ",")

	if len(*o) == 1 && (*o)[0] == "" {
		*o = []string{}
	}

	return nil
}

func (o StringArray) Value() (driver.Value, error) {
	if len(o) == 0 {
		return "{}", nil
	}
	var output string
	for _, val := range o {
		output += fmt.Sprintf("\"%s\",", val)
	}
	return fmt.Sprintf("{%s}", strings.TrimRight(output, ",")), nil
}

func (StringArray) GormDataType() string {
	return "text[]"
}

type Prompt struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Prompt      string      `json:"prompt"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at,omitempty"`
	Tags        StringArray `json:"tags,omitempty"`
	Public      bool        `json:"public"`
	UserID      string      `json:"user_id"`
	Slug        string      `json:"slug"`
}

func GetPromptByIDOrSlug(id string) (*Prompt, error) {
	prompt := &Prompt{}

	matchUUID, _ := regexp.MatchString(UUIDRegexp, id)
	matchSlug, _ := regexp.MatchString(SlugRegexp, id)

	if !matchUUID && !matchSlug {
		return nil, fmt.Errorf("Invalid identifier")
	}

	var err error

	if matchUUID {
		err = DB.First(prompt, "id = ?", id).Error
	} else {
		err = DB.First(prompt, "slug = ?", id).Error
	}

	if err != nil {
		return nil, err
	}

	return prompt, nil
}

func RetrieveSystemPromptID(systemPromptIdOrSlug *string) (*string, error) {
	var prompt Prompt

	if systemPromptIdOrSlug == nil {
		return nil, nil
	}

	matchUUID, _ := regexp.MatchString(UUIDRegexp, *systemPromptIdOrSlug)
	matchSlug, _ := regexp.MatchString(SlugRegexp, *systemPromptIdOrSlug)

	if !matchUUID && !matchSlug {
		return nil, fmt.Errorf("Invalid identifier")
	}

	if matchUUID {
		return systemPromptIdOrSlug, nil
	}

	err := DB.Table("prompts").Select("id").Where("slug = ?", systemPromptIdOrSlug).First(&prompt).Error

	if err != nil {
		return nil, err
	}

	return &prompt.ID, nil
}
