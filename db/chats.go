package db

type Chat struct {
	ID             string        `json:"id,omitempty"`
	UserID         string        `json:"user_id"`
	SystemPrompt   *string       `json:"system_prompt"`
	SystemPromptID *string       `json:"system_prompt_id"`
	ChatMessages   []ChatMessage `json:"chat_messages,omitempty"`
	Name           *string       `json:"name"`
}

func (Chat) TableName() string {
	return "chats"
}

func GetChatByID(id string) (*Chat, error) {
	var result *Chat

	err := DB.First(&result, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func CreateChat(userID string, systemPrompt *string, SystemPromptID *string, name *string) (*Chat, error) {
	var result *Chat

	err := DB.Raw("INSERT INTO chats (user_id, system_prompt, system_prompt_id, name) VALUES (?::uuid, ?, ?, ?) RETURNING *", userID, systemPrompt, SystemPromptID, name).
		Scan(&result).
		Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func ListChats(userID string) ([]Chat, error) {
	var result []Chat

	err := DB.Raw(`
	SELECT c.*, cm.content AS latest_message_content, cm.created_at AS latest_message_created_at
	FROM chats c
	LEFT JOIN LATERAL (
		SELECT cm.content, cm.created_at
		FROM chat_messages cm
		WHERE cm.chat_id = c.id
		ORDER BY cm.created_at DESC
		LIMIT 1
	) cm ON true
	WHERE c.user_id = ?
	`, userID).Scan(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func DeleteChat(userID string, id string) error {
	err := DB.Exec("DELETE FROM chats WHERE id = ? AND user_id = ?", id, userID).Error

	return err
}

func UpdateChat(userID string, id string, name string) (*Chat, error) {
	var result *Chat

	err := DB.Raw("UPDATE chats SET name = ? WHERE id = ? AND user_id = ? RETURNING *", name, id, userID).Scan(&result).Error
	if err != nil {
		return nil, err
	}

	return result, err
}

type ChatMessage struct {
	ID            *string `json:"id"`
	ChatID        string  `json:"chat_id"`
	IsUserMessage bool    `json:"is_user_message"`
	Content       string  `json:"content"`
	CreatedAt     string  `json:"created_at"`
}

func GetChatMessages(userID string, chatID string) ([]ChatMessage, error) {
	results := make([]ChatMessage, 0)

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
