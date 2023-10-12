package db

type PromptUses struct {
	UserId   string `json:"user_id"`
	PromptId string `json:"prompt_id"`
}

func (PromptUses) TableName() string {
	return "prompts_uses"
}

func AddPromptUse(user_id string, prompt_id string) (*PromptUses, error) {
	var result *PromptUses

	err := DB.Create(&PromptUses{UserId: user_id, PromptId: prompt_id}).Error

	if err != nil {
		return nil, err
	}

	return result, nil
}

func GetPromptUsesByPromptId(prompt_id string) ([]PromptUses, error) {
	var results []PromptUses

	err := DB.Where("prompt_id = ?", prompt_id).Find(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

func GetPromptUsesByUserId(user_id string) ([]PromptUses, error) {
	var results []PromptUses

	err := DB.Where("user_id = ?", user_id).Find(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

func GetPromptUsesByUserIdAndPromptId(user_id string, prompt_id string) (*PromptUses, error) {
	var result PromptUses

	err := DB.Where("prompt_id = ?", prompt_id).Where("user_id = ?", user_id).First(&result).Error

	if err != nil {
		return nil, err
	}

	return &result, nil
}
