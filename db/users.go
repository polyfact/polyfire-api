package db

import (
	"errors"
	"sync"
)

type User struct {
	Id      string `json:"id"`
	Version int    `json:"version,omitempty"`
}

func getUserDBVersion(auth_id string) (*User, error) {
	client, err := CreateClient()
	if err != nil {
		return nil, err
	}

	var results []User

	_, err = client.From("auth_users").Select("*", "exact", false).Eq("id", auth_id).ExecuteTo(&results)

	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}

	return &results[0], nil
}

var (
	UnknownUserId     = errors.New("Unknown user id")
	DBVersionMismatch = errors.New("DB version mismatch")
	DBError           = errors.New("Database error")
)

type RateLimitStatus string

var (
	RateLimitStatusReached        = RateLimitStatus("rate_limit_reached")
	RateLimitStatusProjectReached = RateLimitStatus("project_rate_limit_reached")
	RateLimitStatusOk             = RateLimitStatus("ok")
	RateLimitStatusNone           = RateLimitStatus("")
)

func checkUserRateLimit(user_id string) (RateLimitStatus, error) {
	var userReached bool
	var userRateLimitErr error

	userReached, userRateLimitErr = UserReachedRateLimit(user_id)

	if userRateLimitErr != nil {
		return RateLimitStatusNone, UnknownUserId
	}

	if userReached {
		return RateLimitStatusReached, nil
	}

	return RateLimitStatusOk, nil
}

func CheckDBVersionRateLimit(user_id string, version int) (*AuthUser, RateLimitStatus, error) {
	var wg sync.WaitGroup
	var user *User
	var userErr error
	var rateLimitStatus RateLimitStatus
	var rateLimitErr error
	var devAuthUser AuthUser
	var devAuthUserErr error

	wg.Add(3)

	go func() {
		defer wg.Done()
		user, userErr = getUserDBVersion(user_id)
	}()

	go func() {
		defer wg.Done()
		rateLimitStatus, rateLimitErr = checkUserRateLimit(user_id)
	}()

	go func() {
		defer wg.Done()
		devAuthUser, devAuthUserErr = GetDevAuthUserForUserIDProject(user_id)
	}()

	wg.Wait()

	if userErr != nil {
		return nil, RateLimitStatusNone, DBError
	}

	if user != nil && user.Version != version {
		return nil, RateLimitStatusNone, DBVersionMismatch
	}

	if rateLimitErr != nil {
		return nil, RateLimitStatusNone, rateLimitErr
	}

	if devAuthUserErr != nil {
		return nil, RateLimitStatusNone, devAuthUserErr
	}

	if devAuthUser.Usage >= devAuthUser.RateLimit {
		return nil, RateLimitStatusProjectReached, nil
	}

	return &devAuthUser, rateLimitStatus, nil
}
