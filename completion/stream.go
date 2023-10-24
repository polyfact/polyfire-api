package completion

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	router "github.com/julienschmidt/httprouter"
	providers "github.com/polyfire/api/llm/providers"
	utils "github.com/polyfire/api/utils"
	webrequest "github.com/polyfire/api/web_request"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // For now, allow all origins

	// CheckOrigin: func(r *http.Request) bool {
	// 	allowedOrigins := []string{"http://localhost:3000"}
	// 	origin := r.Header["Origin"][0]
	// 	for _, allowedOrigin := range allowedOrigins {
	// 		if origin == allowedOrigin {
	// 			return true
	// 		}
	// 	}
	// 	return false
	// },

}

func Stream(w http.ResponseWriter, r *http.Request, _ router.Params) {
	record := r.Context().Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)
	user_id := r.Context().Value(utils.ContextKeyUserID).(string)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.RespondError(w, record, "communication_error")
		return
	}
	defer conn.Close()

	messageType, p, err := conn.ReadMessage()
	if err != nil {
		utils.RespondErrorStream(conn, record, "read_message_error")
		return
	}

	if messageType != websocket.TextMessage {
		utils.RespondErrorStream(conn, record, "invalid_message_type")
		return
	}

	recordEventRequest := r.Context().Value(utils.ContextKeyRecordEventRequest).(utils.RecordRequestFunc)

	record = func(response string, props ...utils.KeyValue) {
		recordEventRequest(string(p), response, user_id, props...)
	}

	var input GenerateRequestBody

	err = json.Unmarshal(p, &input)
	if err != nil {
		utils.RespondErrorStream(conn, record, "invalid_json")
		return
	}

	chan_res, err := GenerationStart(r.Context(), user_id, input)
	if err != nil {
		if err != nil {
			switch err {
			case webrequest.WebsiteExceedsLimit:
				utils.RespondErrorStream(conn, record, "error_website_exceeds_limit")
			case webrequest.WebsitesContentExceeds:
				utils.RespondErrorStream(conn, record, "error_websites_content_exceeds")
			case webrequest.NoContentFound:
				utils.RespondErrorStream(conn, record, "error_no_content_found")
			case webrequest.FetchWebpageError:
				utils.RespondErrorStream(conn, record, "error_fetch_webpage")
			case webrequest.ParseContentError:
				utils.RespondErrorStream(conn, record, "error_parse_content")
			case webrequest.VisitBaseURLError:
				utils.RespondErrorStream(conn, record, "error_visit_base_url")
			case NotFound:
				utils.RespondErrorStream(conn, record, "not_found")
			case UnknownModelProvider:
				utils.RespondErrorStream(conn, record, "invalid_model_provider")
			case RateLimitReached:
				utils.RespondErrorStream(conn, record, "rate_limit_reached")
			case ProjectRateLimitReached:
				utils.RespondErrorStream(conn, record, "project_rate_limit_reached")
			default:
				utils.RespondErrorStream(conn, record, "internal_error")
			}
			return
		}
	}

	result := providers.Result{
		Result:     "",
		TokenUsage: providers.TokenUsage{Input: 0, Output: 0},
	}

	chan_stop := make(chan bool)
	go func() {
		for {
			size, message, _ := conn.ReadMessage()
			if string(message) == "STOP" {
				chan_stop <- true
			}
			if size == -1 {
				break
			}
		}
	}()

	total_result := ""
generation_loop:
	for v := range *chan_res {
		result.Result += v.Result
		if v.TokenUsage.Input != 0 {
			result.TokenUsage.Input = v.TokenUsage.Input
		}
		result.TokenUsage.Output += v.TokenUsage.Output

		if len(v.Resources) > 0 {
			result.Resources = v.Resources
		}
		select {
		case <-chan_stop:
			break generation_loop
		default:
		}

		total_result += v.Result
		if v.Result != "" {
			err = conn.WriteMessage(websocket.TextMessage, []byte(v.Result))
			if err != nil {
				utils.RespondErrorStream(conn, record, "write_message_error")
				return
			}
		}
	}

	if input.Infos {
		infosJSON, err := result.JSON()
		if err != nil {
			utils.RespondErrorStream(conn, record, "invalid_json")
			return
		}

		infos := "[INFOS]:" + string(infosJSON)
		byteMessage := []byte(infos)

		err = conn.WriteMessage(websocket.TextMessage, byteMessage)
		if err != nil {
			utils.RespondErrorStream(conn, record, "write_info_error")
			return
		}
	}

	record(total_result)

	err = conn.WriteMessage(websocket.TextMessage, []byte(""))
	if err != nil {
		utils.RespondErrorStream(conn, record, "write_end_message_error")
		return
	}
}
