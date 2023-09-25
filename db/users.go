package db

import (
	"errors"
	"sync"
)

type UserAuth struct {
	Id          string `json:"id"`
	Version     int    `json:"version,omitempty"`
	OpenAIToken string `json:"openai_token,omitempty"`
	OpenAIOrg   string `json:"openai_org,omitempty"`
}

func getAuthUser(auth_id string) (*UserAuth, error) {
	client, err := CreateClient()
	if err != nil {
		return nil, err
	}

	var results []UserAuth

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

func CheckDBVersionRateLimit(user_id string, version int) (*UserAuth, RateLimitStatus, error) {
	var wg sync.WaitGroup
	var user *UserAuth
	var userErr error
	var rateLimitStatus RateLimitStatus
	var rateLimitErr error

	wg.Add(2)

	go func() {
		defer wg.Done()
		user, userErr = getAuthUser(user_id)
	}()

	go func() {
		defer wg.Done()
		rateLimitStatus, rateLimitErr = checkRateLimit(user_id)
	}()

	wg.Wait()

	if userErr != nil {
		return nil, RateLimitStatusNone, DBError
	}

	if user.Version != version {
		return nil, RateLimitStatusNone, DBVersionMismatch
	}

	if rateLimitErr != nil {
		return nil, RateLimitStatusNone, rateLimitErr
	}

	return user, rateLimitStatus, nil
}
