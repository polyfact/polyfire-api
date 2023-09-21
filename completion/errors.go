package completion

import (
	"errors"
)

var (
	UnknownUserId           error = errors.New("400 Unknown user Id")
	InternalServerError     error = errors.New("500 InternalServerError")
	UnknownModelProvider    error = errors.New("400 Unknown model provider")
	NotFound                error = errors.New("404 Not Found")
	RateLimitReached        error = errors.New("429 Monthly Rate Limit Reached")
	ProjectRateLimitReached error = errors.New("429 Monthly Project Rate Limit Reached")
	ProjectNotPremiumModel  error = errors.New("403 Project Can't Use Premium Models")
	UnknownError            error = errors.New("500 Unknown Error")
)
