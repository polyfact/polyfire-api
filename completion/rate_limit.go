package completion

import (
	"sync"

	"github.com/polyfact/api/db"
)

func CheckRateLimit(user_id string) error {
	var wg sync.WaitGroup
	var userReached bool
	var projectReached bool

	var userRateLimitErr error
	var projectRateLimitErr error

	wg.Add(2)

	go func() {
		defer wg.Done()
		userReached, userRateLimitErr = db.UserReachedRateLimit(user_id)
	}()

	go func() {
		defer wg.Done()
		projectReached, projectRateLimitErr = db.ProjectReachedRateLimit(user_id)
	}()

	wg.Wait()

	if userRateLimitErr != nil {
		return UnknownUserId
	}
	if projectRateLimitErr != nil {
		return UnknownUserId
	}

	if userReached {
		return RateLimitReached
	}
	if projectReached {
		return ProjectRateLimitReached
	}

	return nil
}
