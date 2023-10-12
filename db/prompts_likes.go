package db

import (
	"time"

	"gorm.io/gorm"
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
	CreatedAt time.Time `json:"created_at"`
}

func (PromptLike) TableName() string {
	return "prompts_likes"
}

func GetPromptLikeByUserId(input PromptLikeInput) (*PromptLike, error) {
	var promptLike PromptLike
	if err := DB.Where("user_id = ? AND prompt_id = ?", input.UserId, input.PromptId).First(&promptLike).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &promptLike, nil
}

func AddPromptLike(input PromptLikeInput) (*PromptLike, error) {
	promptLike := PromptLike{
		UserId:   input.UserId,
		PromptId: input.PromptId,
	}

	if err := DB.Create(&promptLike).Error; err != nil {
		return nil, err
	}
	return &promptLike, nil
}

func RemovePromptLike(input PromptLikeInput) error {
	if err := DB.Where("user_id = ? AND prompt_id = ?", input.UserId, input.PromptId).Delete(&PromptLike{}).Error; err != nil {
		return err
	}
	return nil
}
