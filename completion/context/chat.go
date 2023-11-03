package context

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/polyfire/api/db"
)

var CHAT_HISTORY_TEMPLATE = template.Must(
	template.New("chat_history_context").Parse(`Here's the previous conversation history:
==========
{{range .Data}}{{.}}
{{end}}==========
`),
)

type ChatHistoryContext struct {
	Messages []string
}

var CHAT_HISTORY_TEMPLATE_GROWTH = InitContextStructureTemplate(*CHAT_HISTORY_TEMPLATE)

func GetChatHistoryContext(user_id string, chatId string) (*ChatHistoryContext, error) {
	allHistory, err := db.GetChatMessages(user_id, chatId)
	if err != nil {
		return nil, err
	}

	var messages []string
	for _, message := range allHistory {
		if message.IsUserMessage {
			messages = append(messages, fmt.Sprintf("User: %s", message.Content))
		} else {
			messages = append(messages, fmt.Sprintf("AI: %s", message.Content))
		}
	}

	chatHistoryContext := ChatHistoryContext{
		Messages: messages,
	}

	return &chatHistoryContext, nil
}

func (chc *ChatHistoryContext) GetPriority() Priority {
	return IMPORTANT
}

func (chc *ChatHistoryContext) GetOrderIndex() int {
	return 3
}

func (chc *ChatHistoryContext) GetMinimumContextSize() int {
	if len(chc.Messages) == 0 {
		return 0
	}
	return CHAT_HISTORY_TEMPLATE_GROWTH.B + CHAT_HISTORY_TEMPLATE_GROWTH.A + len(chc.Messages[0])
}

func (chc *ChatHistoryContext) GetRecommendedContextSize() int {
	if len(chc.Messages) == 0 {
		return 0
	}
	totalSize := CHAT_HISTORY_TEMPLATE_GROWTH.B
	for i := 0; i < len(chc.Messages); i++ {
		totalSize += CHAT_HISTORY_TEMPLATE_GROWTH.A + len(chc.Messages[i])
	}

	return totalSize
}

type ChatHistoryTemplateData struct {
	Data []string
}

func (spc *ChatHistoryContext) GetContentFittingIn(tokenCount int) string {
	tokenCurrentSize := CHAT_HISTORY_TEMPLATE_GROWTH.B
	var result []string
	for i := 0; i < len(spc.Messages); i++ {
		tokenCurrentSize += CHAT_HISTORY_TEMPLATE_GROWTH.A + len(spc.Messages[i])
		if tokenCurrentSize > tokenCount {
			break
		}
		result = append([]string{spc.Messages[i]}, result...)
	}

	templData := ChatHistoryTemplateData{
		Data: result,
	}

	var resultBuf bytes.Buffer

	if err := CHAT_HISTORY_TEMPLATE.Execute(&resultBuf, templData); err != nil {
		return ""
	}

	return resultBuf.String()
}
