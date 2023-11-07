package db

type Chat struct {
	ID             string        `json:"id,omitempty"`
	UserID         string        `json:"user_id"`
	SystemPrompt   *string       `json:"system_prompt"`
	SystemPromptID *string       `json:"system_prompt_id"`
	ChatMessages   []ChatMessage `json:"chat_messages"`
}

func (Chat) TableName() string {
	return "chats"
}

func GetChatForUser(userID string) ([]Chat, error) {
	var results []Chat

	err := DB.Find(&results, "user_id = ?", userID).Error
	if err != nil {
		return nil, err
	}

	return results, nil
}

func GetChatByID(id string) (*Chat, error) {
	var result *Chat

	err := DB.First(&result, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func CreateChat(userID string, systemPrompt *string, SystemPromptID *string) (*Chat, error) {
	var result *Chat

	err := DB.Raw("INSERT INTO chats (user_id, system_prompt, system_prompt_id) VALUES (?::uuid, ?, ?) RETURNING *", userID, systemPrompt, SystemPromptID).
		Scan(&result).
		Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

type ChatMessage struct {
	ID            *string `json:"id"`
	ChatID        string  `json:"chat_id"`
	IsUserMessage bool    `json:"is_user_message"`
	Content       string  `json:"content"`
	CreatedAt     string  `json:"created_at"`
}

func GetChatMessages(userID string, chatID string) ([]ChatMessage, error) {
	var results = make([]ChatMessage, 0)

	err := DB.Raw("SELECT chat_messages.* FROM chat_messages JOIN chats ON chats.id = chat_messages.chat_id WHERE chats.id = ? AND chats.user_id = ? ORDER BY chat_messages.created_at DESC LIMIT 20", chatID, userID).
		Scan(&results).
		Error
	if err != nil {
		return nil, err
	}

	return results, nil
}

func AddChatMessage(chatID string, isUserMessage bool, content string) error {
	err := DB.Exec(
		"INSERT INTO chat_messages (chat_id, is_user_message, content) VALUES (?, ?, ?)",
		chatID,
		isUserMessage,
		content,
	).Error
	if err != nil {
		return err
	}

	return nil
}
