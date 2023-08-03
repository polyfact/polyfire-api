package completion

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	router "github.com/julienschmidt/httprouter"
	db "github.com/polyfact/api/db"
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
	ressources := &[]db.MatchResult{}

	err = json.Unmarshal(p, &input)
	if err != nil {
		http.Error(w, "400 Bad Request", http.StatusBadRequest)
		return
	}

	chan_res, err := GenerationStart(user_id, input, ressources)

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

	message := ""
	for r := range *chan_res {

		message += r.Result

		if r.Result != "" {
			err = conn.WriteMessage(websocket.TextMessage, []byte(r.Result))
			if err != nil {
				http.Error(w, "500 Internal server error", http.StatusInternalServerError)
				return
			}
		}
	}
	if input.MemoryId != nil && *input.MemoryId != "" && input.Infos {
		if err != nil {
			http.Error(w, "500 Internal server error", http.StatusInternalServerError)
			return
		}

		input := llm.CountTokens(input.Task, "gpt-4")
		oupout := llm.CountTokens(message, "gpt-4")

		tokenUsage := llm.TokenUsage{Input: input, Output: oupout}

		result := llm.Result{
			Result:     message,
			TokenUsage: tokenUsage,
			Ressources: *ressources,
		}

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
