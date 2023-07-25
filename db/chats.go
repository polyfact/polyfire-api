package db

type Chat struct {
	ID            string        `json:"id"`
	UserID        string        `json:"user_id"`
	chat_messages []ChatMessage `json:"chat_messages"`
}

func GetChatIds(userId string) ([]Chat, error) {
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

type ChatMessage struct {
	ID            string `json:"id"`
	ChatID        string `json:"chat_id"`
	IsUserMessage bool   `json:"is_user_message"`
	Content       string `json:"content"`
}

func GetChatMessages(userId string, chatId string) (Chat, error) {
	client, err := CreateClient()
	if err != nil {
		panic(err)
	}

	var results Chat

	str, _, err := client.From("chats").
		Select("*, chat_messages(*)", "exact", false).
		Single().
		Eq("id", chatId).
		Eq("user_id", userId).
		ExecuteString()

	panic(str)
	// if err != nil {
	// 	panic(err)
	// 	// return , err
	// }

	return results, nil
}
