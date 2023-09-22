package db

import (
	"errors"
	"sync"
)

type UserAuth struct {
	Id      string `json:"id"`
	Version int    `json:"version,omitempty"`
}

func getVersionForUser(user_id string) (int, error) {
	client, err := CreateClient()
	if err != nil {
		return 0, err
	}

	var results []UserAuth

	_, err = client.From("auth_users").Select("*", "exact", false).Eq("id", user_id).ExecuteTo(&results)

	if err != nil {
		return 0, err
	}

	if len(results) == 0 {
		return 0, nil
	}

	return results[0].Version, nil
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

func checkRateLimit(user_id string) (RateLimitStatus, error) {
	var wg sync.WaitGroup
	var userReached bool
	var projectReached bool

	var userRateLimitErr error
	var projectRateLimitErr error

	wg.Add(2)

	go func() {
		defer wg.Done()
		userReached, userRateLimitErr = UserReachedRateLimit(user_id)
	}()

	go func() {
		defer wg.Done()
		projectReached, projectRateLimitErr = ProjectReachedRateLimit(user_id)
	}()

	wg.Wait()

	if userRateLimitErr != nil {
		return RateLimitStatusNone, UnknownUserId
	}
	if projectRateLimitErr != nil {
		return RateLimitStatusNone, UnknownUserId
	}

	if userReached {
		return RateLimitStatusReached, nil
	}
	if projectReached {
		return RateLimitStatusProjectReached, nil
	}

	return RateLimitStatusOk, nil
}

func CheckDBVersionRateLimit(user_id string, version int) (RateLimitStatus, error) {
	var wg sync.WaitGroup
	var dbVersion int
	var dbVersionErr error
	var rateLimitStatus RateLimitStatus
	var rateLimitErr error

	wg.Add(2)

	go func() {
		defer wg.Done()
		dbVersion, dbVersionErr = getVersionForUser(user_id)
	}()

	go func() {
		defer wg.Done()
		rateLimitStatus, rateLimitErr = checkRateLimit(user_id)
	}()

	wg.Wait()

	if dbVersionErr != nil {
		return RateLimitStatusNone, DBError
	}

	if dbVersion != version {
		return RateLimitStatusNone, DBVersionMismatch
	}

	if rateLimitErr != nil {
		return RateLimitStatusNone, rateLimitErr
	}

	return rateLimitStatus, nil
}
