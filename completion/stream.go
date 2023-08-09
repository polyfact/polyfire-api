package completion

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	router "github.com/julienschmidt/httprouter"
	providers "github.com/polyfact/api/llm/providers"
	utils "github.com/polyfact/api/utils"
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
	user_id := r.Context().Value("user_id").(string)
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		utils.RespondError(w, "communication_error")
		return
	}
	defer conn.Close()

	messageType, p, err := conn.ReadMessage()
	if err != nil {
		utils.RespondError(w, "read_message_error")
		return
	}

	if messageType != websocket.TextMessage {
		utils.RespondError(w, "invalid_message_type")
		return
	}

	var input GenerateRequestBody

	err = json.Unmarshal(p, &input)
	if err != nil {
		utils.RespondError(w, "invalid_json")
		return
	}

	chan_res, err := GenerationStart(user_id, input)

	if err != nil {
		switch err {
		case NotFound:
			utils.RespondError(w, "not_found_error")
		case UnknownModelProvider:
			utils.RespondError(w, "unknown_model_provider")
		default:
			utils.RespondError(w, "generation_error")
		}
		return
	}

	result := providers.Result{
		Result:     "",
		TokenUsage: providers.TokenUsage{Input: 0, Output: 0},
	}

	for v := range *chan_res {
		result.Result += v.Result
		result.TokenUsage.Input += v.TokenUsage.Input
		result.TokenUsage.Output += v.TokenUsage.Output

		if len(v.Ressources) > 0 {
			result.Ressources = v.Ressources
		}

		if v.Result != "" {
			err = conn.WriteMessage(websocket.TextMessage, []byte(v.Result))
			if err != nil {
				utils.RespondError(w, "write_message_error")
				return
			}
		}
	}

	if input.MemoryId != nil && *input.MemoryId != "" && input.Infos {
		infosJSON, err := json.Marshal(result)

		infos := "[INFOS]:" + string(infosJSON)
		byteMessage := []byte(infos)

		err = conn.WriteMessage(websocket.TextMessage, byteMessage)
		if err != nil {
			utils.RespondError(w, "write_info_error")
			return
		}
	}

	err = conn.WriteMessage(websocket.TextMessage, []byte(""))
	if err != nil {
		utils.RespondError(w, "write_end_message_error")
		return
	}
}
