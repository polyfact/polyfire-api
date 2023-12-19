package auth

import (
	"encoding/json"
	"net/http"

	router "github.com/julienschmidt/httprouter"
	"github.com/polyfire/api/db"
	"github.com/polyfire/api/utils"
)

type PaymentLinkResponse struct {
	url string `json:"url"`
}

func PaymentLink(w http.ResponseWriter, r *http.Request, _ router.Params) {
	record := r.Context().Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)
	projectID := r.Context().Value(utils.ContextKeyProjectID).(string)

	project, err := db.GetProjectByID(projectID)
	if err != nil {
		utils.RespondError(w, record, "project_retrieval_error")
		return
	}

	result := PaymentLinkResponse{
		url: project.StripePaymentLink,
	}

	response, _ := json.Marshal(&result)
	record(string(response))

	_ = json.NewEncoder(w).Encode(result)
}
