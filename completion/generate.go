package completion

import (
	"encoding/json"
	"errors"
	"net/http"

	router "github.com/julienschmidt/httprouter"
	db "github.com/polyfact/api/db"
	llm "github.com/polyfact/api/llm"
	providers "github.com/polyfact/api/llm/providers"
	memory "github.com/polyfact/api/memory"
	posthog "github.com/polyfact/api/posthog"
	utils "github.com/polyfact/api/utils"
)

type GenerateRequestBody struct {
	Task         string    `json:"task"`
	Provider     string    `json:"provider,omitempty"`
	MemoryId     *string   `json:"memory_id,omitempty"`
	ChatId       *string   `json:"chat_id,omitempty"`
	Stop         *[]string `json:"stop,omitempty"`
	Stream       bool      `json:"stream,omitempty"`
	Infos        bool      `json:"infos,omitempty"`
	PromptId     *string   `json:"prompt_id,omitempty"`
	SystemPrompt *string   `json:"system_prompt,omitempty"`
}

var (
	InternalServerError  error = errors.New("500 InternalServerError")
	UnknownModelProvider error = errors.New("400 Unknown model provider")
	NotFound             error = errors.New("404 Not Found")
	RateLimitReached     error = errors.New("429 Monthly Rate Limit Reached")
)

func GenerationStart(user_id string, input GenerateRequestBody) (*chan providers.Result, error) {
	result := make(chan providers.Result)
	context_completion := ""

	ressources := []db.MatchResult{}

	if input.MemoryId != nil && len(*input.MemoryId) > 0 {
		results, err := memory.Embedder(user_id, *input.MemoryId, input.Task)
		if err != nil {
			return nil, InternalServerError
		}
		if input.Infos {
			ressources = results
		}
		context_completion, err = utils.FillContext(results)

		if err != nil {
			return nil, InternalServerError
		}

	}

	reached, err := db.UserReachedRateLimit(user_id)
	if err != nil {
		return nil, InternalServerError
	}
	if reached {
		return nil, RateLimitReached
	}

	callback := func(model_name string, input_count int, output_count int) {
		db.LogRequests(user_id, model_name, input_count, output_count, "completion")
		posthog.GenerateEvent(user_id, model_name, input_count, output_count)
	}

	if input.Provider == "" {
		input.Provider = "openai"
	}

	provider, err := llm.NewProvider(input.Provider)
	if err == llm.ErrUnknownModel {
		return nil, UnknownModelProvider
	}

	if err != nil {
		return nil, InternalServerError
	}
	var system_prompt string = ""

	if input.PromptId != nil {
		p, err := db.GetPromptById(*input.PromptId)
		db.UpdatePrompt(*input.PromptId, db.PromptUpdate{Use: p.Use + 1})
		if err != nil {
			return nil, NotFound
		}
		system_prompt = p.Prompt
	}

	if input.ChatId != nil && len(*input.ChatId) > 0 {
		chat, err := db.GetChatById(*input.ChatId)
		if err != nil {
			return nil, InternalServerError
		}

		if chat == nil || chat.UserID != user_id {
			return nil, NotFound
		}

		allHistory, err := db.GetChatMessages(user_id, *input.ChatId)
		if err != nil {
			return nil, InternalServerError
		}

		chatHistory := utils.CutChatHistory(allHistory, 1000)

		if chat.SystemPrompt != nil {
			system_prompt = *(chat.SystemPrompt)
		}

		prompt := FormatPrompt(context_completion+"\n"+system_prompt, chatHistory, input.Task)

		err = db.AddChatMessage(chat.ID, true, input.Task)
		if err != nil {
			return nil, InternalServerError
		}

		pre_result := provider.Generate(prompt, &callback, &providers.ProviderOptions{StopWords: &[]string{"AI:", "Human:"}})

		go func() {
			defer close(result)
			total_result := ""
			for v := range pre_result {
				if input.MemoryId != nil && *input.MemoryId != "" && input.Infos && len(ressources) > 0 {
					v.Ressources = ressources
				}

				total_result += v.Result
				result <- v
			}
			err = db.AddChatMessage(chat.ID, false, total_result)
		}()

	} else {

		// Warning: Check if there is a better way to do this to avoid useless parameter:
		if input.PromptId == nil && input.SystemPrompt != nil {
			system_prompt = *(input.SystemPrompt)
		}

		prompt := context_completion + "\n" + system_prompt + "\n" + input.Task

		if input.Stop != nil {
			result = provider.Generate(prompt, &callback, &providers.ProviderOptions{StopWords: input.Stop})
		} else {
			result = provider.Generate(prompt, &callback, nil)
		}
	}

	return &result, nil
}

func Generate(w http.ResponseWriter, r *http.Request, _ router.Params) {
	user_id := r.Context().Value("user_id").(string)

	if len(r.Header["Content-Type"]) == 0 || r.Header["Content-Type"][0] != "application/json" {
		utils.RespondError(w, "invalid_content_type")
		return
	}

	var input GenerateRequestBody

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		utils.RespondError(w, "invalid_json")
		return
	}

	res_chan, err := GenerationStart(user_id, input)
	if err != nil {
		switch err {
		case NotFound:
			utils.RespondError(w, "not_found")
		case UnknownModelProvider:
			utils.RespondError(w, "invalid_model_provider")
		case RateLimitReached:
			utils.RespondError(w, "rate_limit_reached")
		default:
			utils.RespondError(w, "internal_error")
		}
		return
	}

	result := providers.Result{
		Result:     "",
		TokenUsage: providers.TokenUsage{Input: 0, Output: 0},
	}

	for v := range *res_chan {
		result.Result += v.Result
		result.TokenUsage.Input = v.TokenUsage.Input
		result.TokenUsage.Output += v.TokenUsage.Output

		if len(v.Ressources) > 0 {
			result.Ressources = v.Ressources
		}
	}

	w.Header()["Content-Type"] = []string{"application/json"}

	json.NewEncoder(w).Encode(result)
}
