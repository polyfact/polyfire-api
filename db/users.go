package db

import (
	"database/sql"
	"errors"
)

var (
	ErrUnknownUserID     = errors.New("Unknown user id")
	ErrDBVersionMismatch = errors.New("DB version mismatch")
	ErrDB                = errors.New("Database error")
)

type RateLimitStatus string

var (
	RateLimitStatusReached        = RateLimitStatus("rate_limit_reached")
	RateLimitStatusProjectReached = RateLimitStatus("project_rate_limit_reached")
	RateLimitStatusOk             = RateLimitStatus("ok")
	RateLimitStatusNone           = RateLimitStatus("")
)

type UserInfos struct {
	DevRateLimit         int         `json:"dev_rate_limit"`
	DevUsage             int         `json:"dev_usage"`
	ProjectUserRateLimit *int        `json:"project_user_rate_limit"`
	ProjectUserUsage     int         `json:"project_user_usage"`
	Version              int         `json:"version"`
	DevAuthID            string      `json:"dev_auth_id"`
	OpenaiToken          string      `json:"openai_token"` // Somehow this is case sensitive, don't change to OpenAI
	OpenaiOrg            string      `json:"openai_org"`
	ElevenlabsToken      string      `json:"elevenlabs_token"` // Same here, don't change to ElevenLabs
	ReplicateToken       string      `json:"replicate_token"`
	AuthorizedDomains    StringArray `json:"authorized_domains"`
	ProjectID            string      `json:"project_id"`
}

func getUserInfos(userID string) (*UserInfos, error) {
	var userInfos UserInfos

	err := DB.Raw(`
		SELECT
			COALESCE(dev_users.rate_limit, 50000000) as dev_rate_limit,
			COALESCE((SELECT SUM(credits) FROM get_logs_per_projects(dev_users.id, now()::timestamp, (now() - interval '1' month)::timestamp)), 0) as dev_usage,
			project_users.version as version,
			dev_users.id as dev_auth_id,
			dev_users.openai_token as openai_token,
			dev_users.openai_org as openai_org,
			dev_users.replicate_token as replicate_token,
			dev_users.elevenlabs_token as elevenlabs_token,
			projects.authorized_domains as authorized_domains,
			COALESCE(project_users.monthly_credit_rate_limit, default_monthly_credit_rate_limit) as project_user_rate_limit,
			get_monthly_credit_usage(project_users.id::text) as project_user_usage,
			projects.id as project_id
		FROM project_users
		JOIN projects ON project_users.project_id = projects.id
		JOIN auth_users as dev_users ON dev_users.id::text = projects.auth_id::text
		WHERE project_users.id = @id
		LIMIT 1
	`, sql.Named("id", userID)).Scan(&userInfos).Error
	if err != nil {
		return nil, err
	}

	return &userInfos, nil
}

func CheckDBVersionRateLimit(userID string, version int) (*UserInfos, RateLimitStatus, error) {
	userInfos, err := getUserInfos(userID)
	if err != nil {
		return nil, RateLimitStatusNone, err
	}

	if userInfos == nil {
		return nil, RateLimitStatusNone, ErrUnknownUserID
	}

	if userInfos.Version != version {
		return nil, RateLimitStatusNone, ErrDBVersionMismatch
	}

	if userInfos.DevUsage >= userInfos.DevRateLimit {
		return userInfos, RateLimitStatusProjectReached, nil
	}

	if userInfos.ProjectUserRateLimit != nil && userInfos.ProjectUserUsage >= *userInfos.ProjectUserRateLimit {
		return userInfos, RateLimitStatusReached, nil
	}

	return userInfos, RateLimitStatusOk, nil
}

type RefreshToken struct {
	RefreshToken         string `json:"refresh_token"`
	RefreshTokenSupabase string `json:"refresh_token_supabase"`
	ProjectID            string `json:"project_id"`
}

func CreateRefreshToken(refreshToken string, refreshTokenSupabase string, projectID string) error {
	err := DB.Exec(`
		INSERT INTO refresh_tokens (refresh_token, refresh_token_supabase, project_id)
		VALUES (@refresh_token, @refresh_token_supabase, @project_id)
	`, sql.Named("refresh_token", refreshToken), sql.Named("refresh_token_supabase", refreshTokenSupabase), sql.Named("project_id", projectID)).Error
	if err != nil {
		return err
	}

	return nil
}

func GetAndDeleteRefreshToken(refreshToken string) (*RefreshToken, error) {
	var refreshTokenStruct []RefreshToken

	err := DB.Raw(`
		DELETE FROM refresh_tokens
		WHERE refresh_token = @refresh_token
		RETURNING refresh_token, refresh_token_supabase, project_id
	`, sql.Named("refresh_token", refreshToken)).Scan(&refreshTokenStruct).Error
	if err != nil {
		return nil, err
	}

	if len(refreshTokenStruct) == 0 {
		return nil, errors.New("Invalid refresh token")
	}

	return &refreshTokenStruct[0], nil
}

func GetDevEmail(projectID string) (string, error) {
	var devEmail []string

	err := DB.Raw(`SELECT get_dev_email_project_id(@project_id)`, sql.Named("project_id", projectID)).Scan(&devEmail).Error
	if err != nil {
		return "", err
	}

	if len(devEmail) == 0 {
		return "", errors.New("Invalid project id")
	}

	return devEmail[0], nil
}
