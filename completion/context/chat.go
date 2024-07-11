package context

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	database "github.com/polyfire/api/db"
	"github.com/polyfire/api/utils"
)

var chatHistoryTemplate = template.Must(
	template.New("chat_history_context").Parse(`Here's the previous conversation history:
==========
{{range .Data}}{{.}}
{{end}}
`),
)

type ChatHistoryContext struct {
	Messages []string
}

var chatHistoryTemplateGrowth = InitContextStructureTemplate(*chatHistoryTemplate)

func GetChatHistoryContext(
	ctx context.Context,
	userID string,
	chatID string,
) (*ChatHistoryContext, error) {
	db := ctx.Value(utils.ContextKeyDB).(database.Database)
	allHistory, err := db.GetChatMessages(userID, chatID, true, 20, 0)
	if err != nil {
		return nil, err
	}

	var messages []string
	for _, message := range allHistory {
		if strings.TrimSpace(message.Content) != "" {
			if message.IsUserMessage {
				messages = append(messages, fmt.Sprintf("User:\n%s", message.Content))
			} else {
				messages = append(messages, fmt.Sprintf("You:\n%s", message.Content))
			}
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
	return chatHistoryTemplateGrowth.B + chatHistoryTemplateGrowth.A + len(chc.Messages[0])
}

func (chc *ChatHistoryContext) GetRecommendedContextSize() int {
	if len(chc.Messages) == 0 {
		return 0
	}
	totalSize := chatHistoryTemplateGrowth.B
	for i := 0; i < len(chc.Messages); i++ {
		totalSize += chatHistoryTemplateGrowth.A + len(chc.Messages[i])
	}

	return totalSize
}

type ChatHistoryTemplateData struct {
	Data []string
}

func (chc *ChatHistoryContext) GetContentFittingIn(tokenCount int) string {
	tokenCurrentSize := chatHistoryTemplateGrowth.B
	var result []string
	for i := 0; i < len(chc.Messages); i++ {
		tokenCurrentSize += chatHistoryTemplateGrowth.A + len(chc.Messages[i])
		if tokenCurrentSize > tokenCount {
			break
		}
		result = append([]string{chc.Messages[i]}, result...)
	}

	templData := ChatHistoryTemplateData{
		Data: result,
	}

	var resultBuf bytes.Buffer

	if err := chatHistoryTemplate.Execute(&resultBuf, templData); err != nil {
		return ""
	}

	return resultBuf.String()
}
