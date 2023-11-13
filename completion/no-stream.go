package completion

import (
	"encoding/json"
	"net/http"

	router "github.com/julienschmidt/httprouter"
	options "github.com/polyfire/api/llm/providers/options"
	utils "github.com/polyfire/api/utils"
	webrequest "github.com/polyfire/api/web_request"
)

func ReturnErrors(w http.ResponseWriter, record utils.RecordFunc, err error) {
	switch err {
	case webrequest.ErrWebsiteExceedsLimit:
		utils.RespondError(w, record, "error_website_exceeds_limit")
	case webrequest.ErrWebsitesContentExceeds:
		utils.RespondError(w, record, "error_websites_content_exceeds")
	case webrequest.ErrFetchWebpage:
		utils.RespondError(w, record, "error_fetch_webpage")
	case webrequest.ErrParseContent:
		utils.RespondError(w, record, "error_parse_content")
	case webrequest.ErrVisitBaseURL:
		utils.RespondError(w, record, "error_visit_base_url")
	case ErrNotFound:
		utils.RespondError(w, record, "not_found")
	case ErrUnknownModelProvider:
		utils.RespondError(w, record, "invalid_model_provider")
	case ErrRateLimitReached:
		utils.RespondError(w, record, "rate_limit_reached")
	case ErrProjectRateLimitReached:
		utils.RespondError(w, record, "project_rate_limit_reached")
	default:
		utils.RespondError(w, record, "internal_error")
	}
}

func Generate(w http.ResponseWriter, r *http.Request, _ router.Params) {
	userID := r.Context().Value(utils.ContextKeyUserID).(string)
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

	resChan, err := GenerationStart(r.Context(), userID, input)
	if err != nil {
		ReturnErrors(w, record, err)
		return
	}

	result := options.Result{
		Result:     "",
		TokenUsage: options.TokenUsage{Input: 0, Output: 0},
	}

	inputTokens := 0

	for v := range *resChan {
		result.Result += v.Result
		if inputTokens == 0 && v.TokenUsage.Input > 0 {
			inputTokens = v.TokenUsage.Input
			result.TokenUsage.Input = v.TokenUsage.Input
		}
		result.TokenUsage.Output += v.TokenUsage.Output

		if len(v.Resources) > 0 || v.Warnings != nil && len(v.Warnings) > 0 {
			result.Resources = v.Resources
			result.Warnings = v.Warnings
		}

		if v.Err != "" {
			result.Err = v.Err
		}
	}

	w.Header()["Content-Type"] = []string{"application/json"}

	response, _ := result.JSON()
	recordProps := make([]utils.KeyValue, 0)
	if input.SystemPromptID != nil {
		recordProps = append(recordProps, utils.KeyValue{Key: "PromptID", Value: *input.SystemPromptID})
	}
	record(string(response), recordProps...)

	_, _ = w.Write(response)
}
