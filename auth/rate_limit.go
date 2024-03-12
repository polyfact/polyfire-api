package auth

import (
	"encoding/json"
	"net/http"

	router "github.com/julienschmidt/httprouter"
	"github.com/polyfire/api/utils"
)

type UserRateLimitResponse struct {
	Usage     int64  `json:"usage"`
	RateLimit *int64 `json:"rate_limit"`
}

func UserRateLimit(w http.ResponseWriter, r *http.Request, _ router.Params) {
	record := r.Context().Value(utils.ContextKeyRecordEvent).(utils.RecordFunc)
	usage := r.Context().Value(utils.ContextKeyProjectUserUsage).(int64)
	rateLimit := r.Context().Value(utils.ContextKeyProjectUserRateLimit).(*int64)

	result := UserRateLimitResponse{
		Usage:     usage,
		RateLimit: rateLimit,
	}

	response, _ := json.Marshal(&result)
	record(string(response))

	_ = json.NewEncoder(w).Encode(result)
}
