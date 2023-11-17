package completion

import (
	"context"

	"github.com/polyfire/api/db"
	"github.com/polyfire/api/utils"
)

func CheckRateLimit(ctx context.Context) error {
	rateLimitStatus := ctx.Value(utils.ContextKeyRateLimitStatus)

	if rateLimitStatus == db.RateLimitStatusOk {
		return nil
	}

	if rateLimitStatus == db.RateLimitStatusReached {
		return ErrRateLimitReached
	}

	creditsStatus := ctx.Value(utils.ContextKeyCreditsStatus)

	if creditsStatus == db.CreditsStatusUsedUp {
		return ErrCreditsUsedUp
	}

	return ErrUnknownError
}
