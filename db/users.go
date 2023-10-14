package db

import (
	"database/sql"
	"errors"
)

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

type UserInfos struct {
	AuthId            string      `json:"auth_id"`
	DevRateLimit      int         `json:"dev_rate_limit"`
	DevUsage          int         `json:"dev_usage"`
	Version           int         `json:"version"`
	DevAuthId         string      `json:"dev_auth_id"`
	OpenaiToken       string      `json:"openai_token"` // Somehow this is case sensitive, don't change to OpenAI
	OpenaiOrg         string      `json:"openai_org"`
	ReplicateToken    string      `json:"replicate_token"`
	AuthorizedDomains StringArray `json:"authorized_domains"`
}

func getUserInfos(user_id string) (*UserInfos, error) {
	var userInfos UserInfos

	err := DB.Raw(`
		SELECT
			user_users.id as auth_id,
			COALESCE(dev_users.rate_limit, 50000000) as dev_rate_limit,
			COALESCE((SELECT SUM(credits) FROM get_logs_per_projects(dev_users.id, now()::timestamp, (now() - interval '1' month)::timestamp)), 0) as dev_usage,
			CASE WHEN project_users.id = @id THEN project_users.version ELSE user_users.version END as version,
			dev_users.id as dev_auth_id,
			dev_users.openai_token as openai_token,
			dev_users.openai_org as openai_org,
			dev_users.replicate_token as replicate_token,
			projects.authorized_domains as authorized_domains
		FROM
			project_users
		JOIN projects ON project_users.project_id = projects.id
		JOIN auth_users as dev_users ON (project_users.id = @id AND dev_users.id = projects.auth_id::text) OR (project_users.id != @id AND dev_users.id = @id)
		FULL JOIN auth_users as user_users ON user_users.id = project_users.auth_id
		WHERE project_users.id = @id OR user_users.id = @id
		LIMIT 1
	`, sql.Named("id", user_id)).Scan(&userInfos).Error
	if err != nil {
		return nil, err
	}

	return &userInfos, nil
}

func CheckDBVersionRateLimit(user_id string, version int) (*UserInfos, RateLimitStatus, error) {
	userInfos, err := getUserInfos(user_id)
	if err != nil {
		return nil, RateLimitStatusNone, err
	}

	if userInfos == nil {
		return nil, RateLimitStatusNone, UnknownUserId
	}

	if userInfos.Version != version {
		return nil, RateLimitStatusNone, DBVersionMismatch
	}

	if userInfos.DevUsage >= userInfos.DevRateLimit {
		return userInfos, RateLimitStatusProjectReached, nil
	}

	return userInfos, RateLimitStatusOk, nil
}
