package completion

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	router "github.com/julienschmidt/httprouter"
	db "github.com/polyfact/api/db"
	llm "github.com/polyfact/api/llm"
	providers "github.com/polyfact/api/llm/providers"
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
	SystemPromptId *string     `json:"system_prompt_id,omitempty"`
	SystemPrompt   *string     `json:"system_prompt,omitempty"`
	WebRequest     bool        `json:"web,omitempty"`
	Language       *string     `json:"language,omitempty"`
}

func getLanguageCompletion(language *string) string {
	if language != nil && *language != "" {
		return "Answer in " + *language + "."
	}
	return ""
}

func GenerationStart(ctx context.Context, user_id string, input GenerateRequestBody) (*chan providers.Result, error) {
	context_completion := ""
	resources := []db.MatchResult{}

	log.Println("Init provider")
	provider, err := llm.NewProvider(ctx, input.Provider, input.Model)
	if err == llm.ErrUnknownModel {
		return nil, UnknownModelProvider
	}

	if err != nil {
		return nil, InternalServerError
	}

	if provider.DoesFollowRateLimit() {
		log.Println("Check Rate Limit")
		err = CheckRateLimit(ctx)
		if err != nil {
			return nil, err
		}
	}

	callback := func(provider_name string, model_name string, input_count int, output_count int, _completion string, credit *int) {
		if credit != nil && provider.DoesFollowRateLimit() {
			db.LogRequestsCredits(user_id, provider_name, model_name, *credit, input_count, output_count, "completion")
		} else {
			db.LogRequests(
				user_id,
				provider_name,
				model_name,
				input_count,
				output_count,
				"completion",
				provider.DoesFollowRateLimit(),
			)
		}
	}

	chan_memory_res := make(chan *MemoryProcessResult)
	go func() {
		defer close(chan_memory_res)
		memoryResult, err := getMemory(user_id, input.MemoryId, input.Task)
		if err != nil {
			panic(err)
		}

		chan_memory_res <- memoryResult
	}()

	chan_system_prompt := make(chan string)
	go func() {
		defer close(chan_system_prompt)
		system_prompt, err := getSystemPrompt(user_id, input.SystemPromptId, input.SystemPrompt)
		if err != nil {
			panic(err)
		}

		chan_system_prompt <- system_prompt
	}()

	opts := providers.ProviderOptions{}
	if input.Stop != nil {
		opts.StopWords = input.Stop
	}
	if input.Temperature != nil {
		opts.Temperature = input.Temperature
	}

	log.Println("Get ContextTask")
	var prompt string
	var chat_system_prompt *string = nil
	if input.ChatId != nil && len(*input.ChatId) > 0 {
		prompt, chat_system_prompt, err = chatContext(user_id, input.Task, *input.ChatId, &callback, &opts)
		if err != nil {
			return nil, err
		}
	} else if input.WebRequest && input.Provider != "llama" {
		prompt, err = webContext(input.Task, input.Model)
		if err != nil {
			return nil, err
		}
	} else {
		prompt = input.Task
	}

	log.Println("Wait for memory")
	memoryResult := <-chan_memory_res
	if memoryResult != nil {
		context_completion = memoryResult.ContextCompletion
		resources = memoryResult.Resources
	}

	log.Println("Wait for system_prompt")
	system_prompt := <-chan_system_prompt

	if system_prompt == "" && chat_system_prompt != nil {
		system_prompt = *chat_system_prompt
	}

	prompt = context_completion + "\n" + system_prompt + "\n" + getLanguageCompletion(input.Language) + "\n" + prompt

	log.Println("Generate")
	res_chan := provider.Generate(prompt, &callback, &opts)

	result := make(chan providers.Result)

	go func() {
		defer close(result)
		for res := range res_chan {
			result <- res
		}
		result <- providers.Result{Resources: resources}
	}()
	return &result, nil
}

func Generate(w http.ResponseWriter, r *http.Request, _ router.Params) {
	user_id := r.Context().Value(utils.ContextKeyUserID).(string)
	record := r.Context().Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)

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

	res_chan, err := GenerationStart(r.Context(), user_id, input)
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

		if len(v.Resources) > 0 {
			result.Resources = v.Resources
		}
	}

	w.Header()["Content-Type"] = []string{"application/json"}

	response, _ := json.Marshal(&result)
	record(string(response))

	_ = json.NewEncoder(w).Encode(result)
}
