package auth

import (
	"encoding/json"
	"net/http"

	router "github.com/julienschmidt/httprouter"
	db "github.com/polyfact/api/db"
	"github.com/polyfact/api/utils"
)

type UserRateLimitResponse struct {
	Usage     int  `json:"usage"`
	RateLimit *int `json:"rate_limit"`
}

func UserRateLimit(w http.ResponseWriter, r *http.Request, _ router.Params) {
	user_id := r.Context().Value("user_id").(string)

	tokenUsage, err := db.GetUserIdMonthlyTokenUsage(user_id)
	if err != nil {
		utils.RespondError(w, "internal_error")
		return
	}

	projectUser, err := db.GetProjectUserByID(user_id)
	var rateLimit *int = nil
	if projectUser != nil && err == nil {
		rateLimit = projectUser.MonthlyTokenRateLimit
	}

	result := UserRateLimitResponse{
		Usage:     tokenUsage,
		RateLimit: rateLimit,
	}

	json.NewEncoder(w).Encode(result)
}
