package db

import (
	"time"
)

type Chat struct {
	ID             string        `json:"id,omitempty"`
	UserID         string        `json:"user_id"`
	SystemPrompt   *string       `json:"system_prompt"`
	SystemPromptID *string       `json:"system_prompt_id"`
	ChatMessages   []ChatMessage `json:"chat_messages,omitempty"`
	Name           *string       `json:"name"`
}

type ChatWithLatestMessage struct {
	Chat
	LatestMessageContent   *string   `json:"latest_message_content,omitempty"`
	LatestMessageCreatedAt time.Time `json:"latest_message_created_at,omitempty"`
}

func (Chat) TableName() string {
	return "chats"
}

func (ChatWithLatestMessage) TableName() string {
	return "chats"
}

func (db DB) GetChatByID(id string) (*Chat, error) {
	var result *Chat

	err := db.sql.First(&result, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (db DB) CreateChat(
	userID string,
	systemPrompt *string,
	SystemPromptID *string,
	name *string,
) (*Chat, error) {
	var result *Chat

	err := db.sql.Raw("INSERT INTO chats (user_id, system_prompt, system_prompt_id, name) VALUES (?::uuid, ?, ?, ?) RETURNING *", userID, systemPrompt, SystemPromptID, name).
		Scan(&result).
		Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (db DB) ListChats(userID string) ([]ChatWithLatestMessage, error) {
	var result []ChatWithLatestMessage

	err := db.sql.Raw(`
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

func (db DB) DeleteChat(userID string, id string) error {
	err := db.sql.Exec("DELETE FROM chats WHERE id = ? AND user_id = ?", id, userID).Error

	return err
}

func (db DB) UpdateChat(userID string, id string, name string) (*Chat, error) {
	var result *Chat

	err := db.sql.Raw("UPDATE chats SET name = ? WHERE id = ? AND user_id = ? RETURNING *", name, id, userID).
		Scan(&result).
		Error
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

func (ChatMessage) TableName() string {
	return "chat_messages"
}

func (db DB) GetChatMessages(
	userID string,
	chatID string,
	orderByDESC bool,
	limit int,
	offset int,
) ([]ChatMessage, error) {
	var results []ChatMessage

	query := db.sql.
		Select("chat_messages.*").
		Joins("JOIN chats ON chats.id = chat_messages.chat_id").
		Where("chats.id = ? AND chats.user_id = ?", chatID, userID)

	if orderByDESC {
		query = query.Order("chat_messages.created_at DESC")
	} else {
		query = query.Order("chat_messages.created_at ASC")
	}

	query = query.Limit(limit).Offset(offset)

	err := query.Find(&results).Error
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (db DB) AddChatMessage(chatID string, isUserMessage bool, content string) error {
	err := db.sql.Exec(
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
