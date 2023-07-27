package db

import (
	postgrest "github.com/supabase/postgrest-go"
)

type Chat struct {
	ID           string        `json:"id,omitempty"`
	UserID       string        `json:"user_id"`
	SystemPrompt *string       `json:"system_prompt"`
	ChatMessages []ChatMessage `json:"chat_messages"`
}

type ChatInsert struct {
	UserID       string  `json:"user_id"`
	SystemPrompt *string `json:"system_prompt"`
}

func GetChatForUser(userId string) ([]Chat, error) {
	client, err := CreateClient()
	if err != nil {
		return nil, err
	}

	var results []Chat

	_, err = client.From("chats").Select("id", "exact", false).Eq("user_id", userId).ExecuteTo(&results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func GetChatById(id string) (*Chat, error) {
	client, err := CreateClient()
	if err != nil {
		return nil, err
	}

	var result *Chat

	_, err = client.From("chats").Select("*", "exact", false).Eq("id", id).Single().ExecuteTo(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func CreateChat(userId string, systemPrompt *string) (*Chat, error) {
	client, err := CreateClient()
	if err != nil {
		return nil, err
	}

	var result *Chat

	_, err = client.From("chats").Insert(ChatInsert{
		UserID:       userId,
		SystemPrompt: systemPrompt,
	}, false, "", "", "exact").Single().ExecuteTo(&result)

	if err != nil {
		return nil, err
	}

	return result, nil
}

type ChatMessage struct {
	ID            *string `json:"id",omitempty`
	ChatID        string  `json:"chat_id"`
	IsUserMessage bool    `json:"is_user_message"`
	Content       string  `json:"content"`
	CreatedAt     string  `json:"created_at",omitempty`
}

type ChatMessageInsert struct {
	ChatID        string `json:"chat_id"`
	IsUserMessage bool   `json:"is_user_message"`
	Content       string `json:"content"`
}

func GetChatMessages(userId string, chatId string) ([]ChatMessage, error) {
	client, err := CreateClient()
	if err != nil {
		panic(err)
	}

	var result *Chat

	_, err = client.From("chats").
		Select("*, chat_messages(*)", "exact", false).
		Single().
		Eq("id", chatId).
		Eq("user_id", userId).
		Order("created_at", &postgrest.OrderOpts{
			Ascending:    false,
			ForeignTable: "chat_messages",
		}).
		Limit(20, "chat_messages").
		ExecuteTo(&result)

	if err != nil || result == nil {
		return nil, err
	}

	return result.ChatMessages, nil
}

func AddChatMessage(chatId string, isUserMessage bool, content string) error {
	client, err := CreateClient()
	if err != nil {
		return err
	}

	_, _, err = client.From("chat_messages").Insert(ChatMessageInsert{
		ChatID:        chatId,
		IsUserMessage: isUserMessage,
		Content:       content,
	}, false, "", "", "exact").Execute()

	if err != nil {
		return err
	}

	return nil
}
