package completion

import (
	"context"

	"github.com/polyfact/api/db"
	"github.com/polyfact/api/utils"
)

func CheckRateLimit(ctx context.Context) error {
	rateLimitStatus := ctx.Value(utils.ContextKeyRateLimitStatus)

	if rateLimitStatus == db.RateLimitStatusOk {
		return nil
	}

	if rateLimitStatus == db.RateLimitStatusReached {
		return RateLimitReached
	}

	if rateLimitStatus == db.RateLimitStatusProjectReached {
		return ProjectRateLimitReached
	}

	return UnknownError
}
