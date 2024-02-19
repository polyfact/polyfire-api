package completion

import (
	"context"

	"github.com/polyfire/api/db"
	"github.com/polyfire/api/utils"
)

func CheckRateLimit(ctx context.Context) error {
	rateLimitStatus := ctx.Value(utils.ContextKeyRateLimitStatus)

	creditsStatus := ctx.Value(utils.ContextKeyCreditsStatus)

	if rateLimitStatus == db.RateLimitStatusOk && creditsStatus == db.CreditsStatusOk {
		return nil
	}

	if rateLimitStatus == db.RateLimitStatusReached {
		return ErrRateLimitReached
	}

	if creditsStatus == db.CreditsStatusUsedUp {
		return ErrCreditsUsedUp
	}

	return ErrUnknownError
}
