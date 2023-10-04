package db

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
)

type StringArray []string

func (o *StringArray) Scan(src any) error {
	str, ok := src.(string)
	if !ok {
		return errors.New("src value cannot cast to string")
	}
	*o = strings.Split(strings.Trim(str, "{}"), ",")
	return nil
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
	UserId      string      `json:"user_id"`
}

type PromptWithJoin struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Prompt      string      `json:"prompt"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at,omitempty"`
	Likes       int64       `json:"likes"`
	IsLiked     bool        `json:"is_liked"`
	Uses        StringArray `json:"uses,omitempty"`
	Tags        StringArray `json:"tags,omitempty"`
	Public      bool        `json:"public"`
	UserId      string      `json:"user_id"`
}

type PromptInsert struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Prompt      string   `json:"prompt"`
	Tags        []string `json:"tags,omitempty"`
	UserId      string   `json:"user_id"`
	Public      bool     `json:"public"`
}

type PromptUpdate struct {
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	Name        string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	Prompt      string    `json:"prompt,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	Public      bool      `json:"public"`
}

type FilterOperation string

const (
	Eq    FilterOperation = "eq"    // equals
	Neq   FilterOperation = "neq"   // not equals
	Gt    FilterOperation = "gt"    // greater than
	Lt    FilterOperation = "lt"    // less than
	Gte   FilterOperation = "gte"   // greater than or equal
	Lte   FilterOperation = "lte"   // less than or equal
	Like  FilterOperation = "like"  // pattern match
	Ilike FilterOperation = "ilike" // pattern match, case-insensitive
	Cs    FilterOperation = "cs"    // contains
	Is    FilterOperation = "is"    // is null
	In    FilterOperation = "in"    // in
	Fts   FilterOperation = "fts"   // full-text search
	Plfts FilterOperation = "plfts" // phrase full-text search
	Phfts FilterOperation = "phfts" // phrase full-text search
	Wfts  FilterOperation = "wfts"  // web search

)

type SupabaseFilter struct {
	Column    string
	Operation string
	Value     string
}

type SupabaseFilters []SupabaseFilter

func (Prompt) TableName() string {
	return "prompts"
}

func (PromptUpdate) TableName() string {
	return "prompts"
}

func (PromptWithJoin) TableName() string {
	return "prompts"
}

func StringToFilterOperation(op string) (FilterOperation, error) {
	switch FilterOperation(op) {
	case Eq, Neq, Gt, Lt, Gte, Lte, Like, Ilike, Cs, Is, In, Fts, Plfts, Phfts, Wfts:
		return FilterOperation(op), nil
	default:
		return "", fmt.Errorf("invalid filter operation: %s", op)
	}
}

var AllowedColumns = map[string]bool{
	"name":        true,
	"description": true,
	"tags":        true,
}

var usesField = `array(SELECT prompts_uses.created_at FROM prompts_uses WHERE prompts_uses.prompt_id = prompts.id ) as uses`
var likesField = `(SELECT COUNT(*) FROM prompts_likes WHERE prompts_likes.prompt_id = prompts.id) as likes`

var minField = `prompts.id, prompts.name, prompts.description, prompts.use, prompts.tags, prompts.public, prompts.user_id`
var maxField = `prompts.id, prompts.name, prompts.description, prompts.prompt, prompts.created_at, prompts.updated_at, prompts.tags, prompts.public, prompts.user_id`

func GetPromptByIdOrSlug(id string) (*Prompt, error) {
	fieldMappings := map[string]string{
		"uuid": "id",
		"slug": "name",
	}

	query, err := PrepareQueryByStringIdentifier(id, fieldMappings)
	if err != nil {
		return nil, err
	}

	prompt := &Prompt{}

	DB.First(prompt, query, id)

	if DB.Error != nil {
		return nil, DB.Error
	}

	return prompt, nil
}

func getIsLikeField(userId string) string {
	if userId == "" {
		return ""
	}
	return fmt.Sprintf(`EXISTS (SELECT 1 FROM prompts_likes WHERE prompts_likes.prompt_id = prompts.id AND prompts_likes.user_id = '%s') as is_liked`, userId)
}

func buildFields(baseFields []string, userId string) string {
	fields := append([]string(nil), baseFields...)

	if islikeField := getIsLikeField(userId); islikeField != "" {
		fields = append(fields, islikeField)
	}
	return strings.Join(fields, ", ")
}

func getSelectableFields(userId string) string {
	return buildFields([]string{maxField, usesField, likesField}, userId)
}

func getSelectableMinFields(userId string) string {
	return buildFields([]string{minField, usesField, likesField}, userId)
}

func applyAndValidateFilter(sqlQuery *gorm.DB, filter SupabaseFilter, value string) error {
	_, ok := AllowedColumns[filter.Column]
	if !ok {
		return fmt.Errorf("invalid_column")
	}

	op, err := StringToFilterOperation(filter.Operation)
	if err != nil {
		return err
	}

	sqlQuery.Where(fmt.Sprintf("%s %s ?", filter.Column, op), value)

	return nil

}

func GetAllPrompts(filters SupabaseFilters, userId string, public bool, onlyLiked bool) ([]PromptWithJoin, error) {
	var results []PromptWithJoin

	sqlQuery := DB.Model(&PromptWithJoin{}).
		Select(getSelectableMinFields(userId))

	if onlyLiked {
		sqlQuery = sqlQuery.Joins("INNER JOIN prompts_likes ON prompts_likes.prompt_id = prompts.id").
			Where("prompts_likes.user_id = ? AND prompts.user_id != ?", userId, userId)
	}

	for _, filter := range filters {

		value := filter.Value

		if len(value) > 32 {
			return nil, fmt.Errorf("invalid_length_value")
		}

		switch FilterOperation(filter.Operation) {
		case Cs:
			value = "{" + value + "}"
		case Ilike, Like:
			value = "%" + value + "%"
		}

		err := applyAndValidateFilter(sqlQuery, filter, value)

		if err != nil {
			return nil, err
		}
	}

	if public == false {
		sqlQuery = sqlQuery.Where("user_id = ?", userId)
	} else {
		sqlQuery = sqlQuery.Where("public = true")
	}

	err := sqlQuery.Scan(&results).Error
	if err != nil {
		return nil, err
	}

	log.Println(results)

	return results, nil
}

func GetPromptById(id string, userId string) (*PromptWithJoin, error) {
	var prompt PromptWithJoin

	sqlQuery := DB.Model(&PromptWithJoin{}).
		Select(getSelectableFields(userId)).
		Where("prompts.id = ?", id)

	err := sqlQuery.Scan(&prompt).Error
	if err != nil {
		return nil, err
	}

	return &prompt, nil
}

func GetPromptByName(name string, userId string) (*PromptWithJoin, error) {
	var prompt PromptWithJoin

	sqlQuery := DB.Model(&PromptWithJoin{}).
		Select(getSelectableFields(userId)).
		Where("prompts.name = ?", name)

	err := sqlQuery.Scan(&prompt).Error
	if err != nil {
		return nil, err
	}

	return &prompt, nil
}

func CreatePrompt(input PromptInsert) (*Prompt, error) {
	var result Prompt

	err := DB.Create(&result).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func UpdatePrompt(id string, input PromptUpdate, user_id string) (*Prompt, error) {
	var result Prompt

	input.UpdatedAt = time.Now()

	err := DB.Table("prompts").Where("id = ? AND user_id = ?", id, user_id).Updates(input).First(&result).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("failed to find prompt with id: %s", id)
		}
		return nil, err
	}

	return &result, nil
}

func DeletePrompt(id string, user_id string) error {
	result := Prompt{}

	err := DB.Where("id = ? AND user_id = ?", id, user_id).Delete(&result).Error
	if err != nil {
		return err
	}

	if result.ID == "" {
		return fmt.Errorf("failed to delete prompt with id: %s", id)
	}

	return nil
}
