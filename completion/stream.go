package completion

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	db "github.com/polyfact/api/db"
	"github.com/sashabaranov/go-openai"
)

type StreamRequestBody struct {
	Task     string    `json:"task"`
	MemoryId *string   `json:"memory_id,omitempty"`
	ChatId   *string   `json:"chat_id,omitempty"`
	Provider string    `json:"provider,omitempty"`
	Stop     *[]string `json:"stop,omitempty"`
	Stream   bool      `json:"stream,omitempty"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func wsHandler(w http.ResponseWriter, r *http.Request, prompt string, c *func(string, int, int)) (string, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Upgrade error: %v\n", err)
		return "", err
	}

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	ctx := context.Background()

	req := openai.ChatCompletionRequest{
		Model:     openai.GPT3Dot5Turbo,
		MaxTokens: 20,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Lorem ipsum",
			},
		},
		Stream: true,
	}
	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v\n", err)
		return "", err
	}
	defer stream.Close()

	message := ""
	for {
		response, err := stream.Recv()

		if errors.Is(err, io.EOF) {
			fmt.Println("\nStream finished")
			return message, nil
		}

		if err != nil {
			fmt.Printf("\nStream error: %v\n", err)
			return "", err
		}
		message += response.Choices[0].Delta.Content

		fmt.Printf("\nResponse: %v\n", response.Choices[0].Delta.Content)
		err = conn.WriteMessage(websocket.TextMessage, []byte(response.Choices[0].Delta.Content))
		if err != nil {
			return "", err
		}
	}

}

func Stream(w http.ResponseWriter, r *http.Request) {
	user_id := r.Context().Value("user_id").(string)
	var input StreamRequestBody

	// context_completion := ""

	// if input.MemoryId != nil && len(*input.MemoryId) > 0 {
	// 	results, err := memory.Embedder(user_id, *input.MemoryId, input.Task)
	// 	if err != nil {
	// 		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
	// 		return
	// 	}

	// 	context_completion, err = utils.FillContext(results)

	// 	if err != nil {
	// 		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
	// 		return
	// 	}

	// }

	callback := func(model_name string, input_count int, output_count int) {
		db.LogRequests(user_id, model_name, input_count, output_count, "completion")
	}

	if input.Provider == "" {
		input.Provider = "openai"
	} else if input.Provider != "openai" {
		http.Error(w, "Stream usable only with OPENAI", http.StatusBadRequest)
		return
	}

	res, err := wsHandler(w, r, "", &callback)

	if err != nil {
		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
		return
	}

	fmt.Println(res)

	// if input.ChatId != nil && len(*input.ChatId) > 0 {
	// 	chat, err := db.GetChatById(*input.ChatId)
	// 	if err != nil {
	// 		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
	// 		return
	// 	}

	// 	if chat == nil || chat.UserID != user_id {
	// 		http.Error(w, "404 Not found", http.StatusNotFound)
	// 		return
	// 	}

	// 	allHistory, err := db.GetChatMessages(user_id, *input.ChatId)
	// 	if err != nil {
	// 		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
	// 		return
	// 	}

	// 	chatHistory := utils.CutChatHistory(allHistory, 1000)

	// 	var system_prompt string
	// 	if chat.SystemPrompt == nil {
	// 		system_prompt = ""
	// 	} else {
	// 		system_prompt = *(chat.SystemPrompt)
	// 	}

	// 	prompt := FormatPrompt(context_completion+"\n"+system_prompt, chatHistory, input.Task)

	// 	fmt.Println(chat.ID)
	// 	err = db.AddChatMessage(chat.ID, true, input.Task)
	// 	if err != nil {
	// 		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
	// 		fmt.Println(err)
	// 		return
	// 	}

	// 	res, err := wsHandler(w, r, prompt, &callback)

	// 	fmt.Println(res)

	// 	if err != nil {
	// 		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
	// 		return
	// 	}

	// tokenUsage := TokenUsage{Input: 0, Output: 0}
	// tokenUsage.Input += m.model.GetNumTokens(input_prompt)
	// tokenUsage.Output += m.model.GetNumTokens(completion)

	// var result llm.Result = Result{Result: completion, TokenUsage: tokenUsage}

	// err = db.AddChatMessage(chat.ID, false, result.Result)
	// } else {
	// fmt.Println("no chat id")
	// prompt := context_completion + input.Task

	// res, err := wsHandler(w, r, prompt, &callback)

	// if err != nil {
	// 	http.Error(w, "500 Internal server error", http.StatusInternalServerError)
	// 	return
	// }

	// fmt.Println(res)

	// }
}
