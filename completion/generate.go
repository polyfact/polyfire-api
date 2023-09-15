package completion

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"text/template"
	"time"

	router "github.com/julienschmidt/httprouter"
	db "github.com/polyfact/api/db"
	llm "github.com/polyfact/api/llm"
	providers "github.com/polyfact/api/llm/providers"
	memory "github.com/polyfact/api/memory"
	utils "github.com/polyfact/api/utils"
	webrequest "github.com/polyfact/api/web_request"
)

type GenerateRequestBody struct {
	Task           string      `json:"task"`
	Provider       string      `json:"provider,omitempty"`
	Model          *string     `json:"model,omitempty"`
	MemoryId       interface{} `json:"memory_id,omitempty"`
	ChatId         *string     `json:"chat_id,omitempty"`
	Stop           *[]string   `json:"stop,omitempty"`
	Temperature    *float32    `json:"temperature,omitempty"`
	Stream         bool        `json:"stream,omitempty"`
	Infos          bool        `json:"infos,omitempty"`
	SystemPromptId *string     `json:"system_prompt_id,omitempty"`
	SystemPrompt   *string     `json:"system_prompt,omitempty"`
	WebRequest     bool        `json:"web,omitempty"`
	Language       *string     `json:"language,omitempty"`
}

var (
	UnknownUserId           error = errors.New("400 Unknown user Id")
	InternalServerError     error = errors.New("500 InternalServerError")
	UnknownModelProvider    error = errors.New("400 Unknown model provider")
	NotFound                error = errors.New("404 Not Found")
	RateLimitReached        error = errors.New("429 Monthly Rate Limit Reached")
	ProjectRateLimitReached error = errors.New("429 Monthly Project Rate Limit Reached")
	ProjectNotPremiumModel  error = errors.New("403 Project Can't Use Premium Models")
)

func getLanguageCompletion(language *string) string {
	if language != nil && *language != "" {
		return "Answer in " + *language + "."
	}
	return ""
}

type MemoryProcessResult struct {
	Ressources        []db.MatchResult
	ContextCompletion string
}

func getMemory(user_id string, memoryId []string, task string, infos bool) (*MemoryProcessResult, error) {
	if memoryId == nil {
		return nil, nil
	}

	results, err := memory.Embedder(user_id, memoryId, task)
	if err != nil {
		return nil, InternalServerError
	}

	context_completion, err := utils.FillContext(results)
	if err != nil {
		return nil, InternalServerError
	}

	response := &MemoryProcessResult{
		ContextCompletion: context_completion,
	}

	if infos {
		response.Ressources = results
	}

	return response, nil
}

func checkRateLimit(user_id string) error {
	var wg sync.WaitGroup
	var userReached bool
	var projectReached bool

	var userRateLimitErr error
	var projectRateLimitErr error

	wg.Add(2)

	go func() {
		defer wg.Done()
		userReached, userRateLimitErr = db.UserReachedRateLimit(user_id)
	}()

	go func() {
		defer wg.Done()
		projectReached, projectRateLimitErr = db.ProjectReachedRateLimit(user_id)
	}()

	wg.Wait()

	if userRateLimitErr != nil {
		return UnknownUserId
	}
	if projectRateLimitErr != nil {
		return UnknownUserId
	}

	if userReached {
		return RateLimitReached
	}
	if projectReached {
		return ProjectRateLimitReached
	}

	return nil
}

func GenerationStart(user_id string, input GenerateRequestBody) (*chan providers.Result, error) {
	result := make(chan providers.Result)
	context_completion := ""

	ressources := []db.MatchResult{}

	err := checkRateLimit(user_id)

	if err != nil {
		return nil, err
	}

	language_completion := getLanguageCompletion(input.Language)

	var memoryIdArray []string

	if str, ok := input.MemoryId.(string); ok {
		memoryIdArray = append(memoryIdArray, str)
	} else if array, ok := input.MemoryId.([]interface{}); ok {
		for _, item := range array {
			if str, ok := item.(string); ok {
				memoryIdArray = append(memoryIdArray, str)
			}
		}
	}

	memoryResult, err := getMemory(user_id, memoryIdArray, input.Task, input.Infos)
	if err != nil {
		return nil, err
	}

	if memoryResult != nil {
		ressources = memoryResult.Ressources
		context_completion = memoryResult.ContextCompletion
	}

	callback := func(provider_name string, model_name string, input_count int, output_count int) {
		db.LogRequests(user_id, provider_name, model_name, input_count, output_count, "completion")
	}

	provider, err := llm.NewProvider(input.Provider, input.Model)
	if err == llm.ErrUnknownModel {
		return nil, UnknownModelProvider
	}

	// if !provider.UserAllowed(user_id) {
	// 	return nil, ProjectNotPremiumModel
	// }

	if err != nil {
		return nil, InternalServerError
	}
	var system_prompt string = ""

	if input.SystemPromptId != nil && len(*input.SystemPromptId) > 0 {
		p, err := db.GetPromptById(*input.SystemPromptId)

		db.UpdatePromptUse(*input.SystemPromptId, db.PromptUse{Use: p.Use + 1})
		if err != nil || p == nil {
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

		if input.SystemPromptId == nil && chat.SystemPrompt != nil {
			system_prompt = *(chat.SystemPrompt)
		}

		prompt := FormatPrompt(language_completion+"\n"+context_completion+"\n"+system_prompt, chatHistory, input.Task)

		err = db.AddChatMessage(chat.ID, true, input.Task)
		if err != nil {
			return nil, InternalServerError
		}

		pre_result := provider.Generate(prompt, &callback, &providers.ProviderOptions{StopWords: &[]string{"AI:", "Human:"}})

		go func() {
			defer close(result)
			total_result := ""
			for v := range pre_result {
				if memoryIdArray != nil && input.Infos && len(ressources) > 0 {
					v.Ressources = ressources
				}

				total_result += v.Result
				result <- v
			}
			err = db.AddChatMessage(chat.ID, false, total_result)
		}()

	} else if input.WebRequest && input.Provider != "llama" {

		res, err := webrequest.WebRequest(input.Task, input.Model)
		if err != nil {
			return nil, err
		}

		tmpl := `
		Date: {{.Date}}
		Lang: {{.Language}}
		From this request: {{.Task}}
		and using this website content: {{.Content}}
		answer at the above request and don't include appendix information other than the initial request.
		Don't be creative, just be factual.
		Always answer with the same language as the request.
		If you don't know the answer, just say so.
		If website content is not enough, you can use your own knowledge.
		Use All websites content to make your relevant answer.
		`

		data := struct {
			Task     string
			Content  string
			Date     string
			Language string
		}{
			Task:     input.Task,
			Content:  res,
			Date:     time.Now().Format("2006-01-02"),
			Language: language_completion,
		}

		var tpl bytes.Buffer
		t := template.Must(template.New("prompt").Parse(tmpl))

		if err := t.Execute(&tpl, data); err != nil {
			fmt.Println("Error executing template:", err)
			return nil, err
		}

		prompt := tpl.String()
		result = provider.Generate(prompt, &callback, &providers.ProviderOptions{StopWords: input.Stop})

	} else {

		// Warning: Check if there is a better way to do this to avoid useless parameter:
		if input.SystemPromptId == nil && input.SystemPrompt != nil {
			system_prompt = *(input.SystemPrompt)
		}

		prompt := context_completion + "\n" + system_prompt + "\n" + input.Task + "\n" + language_completion

		opts := &providers.ProviderOptions{}
		if input.Stop != nil {
			opts.StopWords = input.Stop
		}
		if input.Temperature != nil {
			opts.Temperature = input.Temperature
		}
		result = provider.Generate(prompt, &callback, opts)
	}

	return &result, nil
}

func Generate(w http.ResponseWriter, r *http.Request, _ router.Params) {
	user_id := r.Context().Value("user_id").(string)
	record := r.Context().Value("recordEvent").(utils.RecordFunc)

	if len(r.Header["Content-Type"]) == 0 || r.Header["Content-Type"][0] != "application/json" {
		utils.RespondError(w, record, "invalid_content_type")
		return
	}

	var input GenerateRequestBody

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		utils.RespondError(w, record, "invalid_json")
		return
	}

	res_chan, err := GenerationStart(user_id, input)
	if err != nil {
		switch err {
		case webrequest.WebsiteExceedsLimit:
			utils.RespondError(w, record, "error_website_exceeds_limit")
		case webrequest.WebsitesContentExceeds:
			utils.RespondError(w, record, "error_websites_content_exceeds")
		case webrequest.NoContentFound:
			utils.RespondError(w, record, "error_no_content_found")
		case webrequest.FetchWebpageError:
			utils.RespondError(w, record, "error_fetch_webpage")
		case webrequest.ParseContentError:
			utils.RespondError(w, record, "error_parse_content")
		case webrequest.VisitBaseURLError:
			utils.RespondError(w, record, "error_visit_base_url")
		case NotFound:
			utils.RespondError(w, record, "not_found")
		case UnknownModelProvider:
			utils.RespondError(w, record, "invalid_model_provider")
		case RateLimitReached:
			utils.RespondError(w, record, "rate_limit_reached")
		case ProjectRateLimitReached:
			utils.RespondError(w, record, "project_rate_limit_reached")
		case ProjectNotPremiumModel:
			utils.RespondError(w, record, "project_not_premium_model")
		default:
			utils.RespondError(w, record, "internal_error", err.Error())
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

	response, _ := json.Marshal(&result)
	record(string(response))

	json.NewEncoder(w).Encode(result)
}
