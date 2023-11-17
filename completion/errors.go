package completion

import (
	"errors"
)

var (
	ErrUnknownUserID           = errors.New("400 Unknown user Id")
	ErrInternalServerError     = errors.New("500 InternalServerError")
	ErrUnknownModelProvider    = errors.New("400 Unknown model provider")
	ErrNotFound                = errors.New("404 Not Found")
	ErrRateLimitReached        = errors.New("429 Monthly Rate Limit Reached")
	ErrCreditsUsedUp           = errors.New("429 Credits Used Up")
	ErrProjectRateLimitReached = errors.New("429 Monthly Project Rate Limit Reached")
	ErrProjectNotPremiumModel  = errors.New("403 Project Can't Use Premium Models")
	ErrUnknownError            = errors.New("500 Unknown Error")
)
