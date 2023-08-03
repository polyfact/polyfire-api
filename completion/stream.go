package completion

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	router "github.com/julienschmidt/httprouter"
	llm "github.com/polyfact/api/llm"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func Stream(w http.ResponseWriter, r *http.Request, _ router.Params) {
	user_id := r.Context().Value("user_id").(string)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "400 Communication Error", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	messageType, p, err := conn.ReadMessage()
	if err != nil {
		http.Error(w, "400 Bad Request", http.StatusBadRequest)
		return
	}

	if messageType != websocket.TextMessage {
		http.Error(w, "400 Bad Request", http.StatusBadRequest)
		return
	}

	var input GenerateRequestBody

	err = json.Unmarshal(p, &input)
	if err != nil {
		http.Error(w, "400 Bad Request", http.StatusBadRequest)
		return
	}

	chan_res, err := GenerationStart(user_id, input)

	if err != nil {
		switch err {
		case NotFound:
			http.Error(w, "404 NotFound", http.StatusNotFound)
		case UnknownModelProvider:
			http.Error(w, "400 Bad Request", http.StatusBadRequest)
		default:
			http.Error(w, "500 Internal server error", http.StatusInternalServerError)
		}
		return
	}

	result := llm.Result{
		Result:     "",
		TokenUsage: llm.TokenUsage{Input: 0, Output: 0},
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
				http.Error(w, "500 Internal server error", http.StatusInternalServerError)
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
			http.Error(w, "500 Internal server error", http.StatusInternalServerError)
			return
		}
	}

	err = conn.WriteMessage(websocket.TextMessage, []byte(""))
	if err != nil {
		http.Error(w, "500 Internal server error", http.StatusInternalServerError)
		return
	}
}
