package db

import (
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrUnknownUserID     = errors.New("Unknown user id")
	ErrDBVersionMismatch = errors.New("DB version mismatch")
	ErrDB                = errors.New("Database error")
	ErrDevNotPremium     = errors.New("Dev not premium")
)

type RateLimitStatus string

var (
	RateLimitStatusReached = RateLimitStatus("rate_limit_reached")
	RateLimitStatusOk      = RateLimitStatus("ok")
	RateLimitStatusNone    = RateLimitStatus("")
)

type CreditsStatus string

var (
	CreditsStatusOk         = CreditsStatus("ok")
	CreditsStatusUsedUp     = CreditsStatus("used_up")
	CreditsStatusNotPremium = CreditsStatus("not_premium")
	CreditsStatusNone       = CreditsStatus("")
)

type UserInfos struct {
	Premium              bool        `json:"premium"`
	Credits              int64         `json:"credits"`
	DevUsage             int         `json:"dev_usage"`
	ProjectUserRateLimit *int64        `json:"project_user_rate_limit"`
	ProjectUserUsage     int64         `json:"project_user_usage"`
	Version              int         `json:"version"`
	DevAuthID            string      `json:"dev_auth_id"`
	OpenaiToken          string      `json:"openai_token"` // Somehow this is case sensitive, don't change to OpenAI
	OpenaiOrg            string      `json:"openai_org"`
	ElevenlabsToken      string      `json:"elevenlabs_token"` // Same here, don't change to ElevenLabs
	ReplicateToken       string      `json:"replicate_token"`
	AuthorizedDomains    StringArray `json:"authorized_domains"`
	ProjectID            string      `json:"project_id"`
	ProjectUserID        string      `json:"project_user_id"`
}

func (db DB) getUserInfos(userID string) (*UserInfos, error) {
	var userInfos UserInfos

	err := db.sql.Raw(`
		SELECT
			dev_users.premium as premium,
			COALESCE((SELECT SUM(credits) FROM get_logs_per_projects(dev_users.id, now()::timestamp, (now() - interval '1' month)::timestamp)), 0) as dev_usage,
			project_users.version as version,
			dev_users.id as dev_auth_id,
			dev_users.credits as credits,
			dev_users.openai_token as openai_token,
			dev_users.openai_org as openai_org,
			dev_users.replicate_token as replicate_token,
			dev_users.elevenlabs_token as elevenlabs_token,
			projects.authorized_domains as authorized_domains,
			CASE
				WHEN projects.dev_rate_limit IS false AND projects.auth_id::text = project_users.auth_id
					THEN NULL
				WHEN project_users.premium IS true
					THEN projects.premium_monthly_credit_rate_limit
				ELSE projects.default_monthly_credit_rate_limit
			END as project_user_rate_limit,
			get_monthly_credit_usage(project_users.id::text) as project_user_usage,
			projects.id as project_id,
			project_users.id as project_user_id
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

func (db DB) CheckDBVersionRateLimit(userID string, version int) (*UserInfos, RateLimitStatus, CreditsStatus, error) {
	userInfos, err := db.getUserInfos(userID)
	if err != nil {
		return nil, RateLimitStatusNone, CreditsStatusNone, err
	}

	if userInfos == nil {
		return nil, RateLimitStatusNone, CreditsStatusNone, ErrUnknownUserID
	}

	if userInfos.Version != version {
		return nil, RateLimitStatusNone, CreditsStatusNone, ErrDBVersionMismatch
	}

	rateLimitStatus := RateLimitStatusOk
	if userInfos.ProjectUserRateLimit != nil && userInfos.ProjectUserUsage >= *userInfos.ProjectUserRateLimit {
		rateLimitStatus = RateLimitStatusReached
	}

	creditsStatus := CreditsStatusOk
	if !userInfos.Premium {
		return nil, rateLimitStatus, CreditsStatusNotPremium, ErrDevNotPremium
	} else if userInfos.Credits <= 0 {
		creditsStatus = CreditsStatusUsedUp
	}

	fmt.Println("devAuthID:", userInfos.DevAuthID)
	fmt.Println("projectUserID:", userInfos.ProjectUserID)
	fmt.Println("projectID:", userInfos.ProjectID)
	fmt.Println("rateLimitStatus:", rateLimitStatus, ", creditsStatus:", creditsStatus)

	return userInfos, rateLimitStatus, creditsStatus, nil
}

func (db DB) RemoveCreditsFromDev(userID string, credits int) error {
	return db.sql.Exec(
		`UPDATE auth_users SET credits = credits - ? WHERE id = (SELECT auth_users.id FROM project_users JOIN projects ON project_users.project_id = projects.id JOIN auth_users ON auth_users.id = projects.auth_id::text WHERE project_users.id = try_cast_uuid(?) LIMIT 1)`,
		credits,
		userID,
	).Error
}

type RefreshToken struct {
	RefreshToken         string `json:"refresh_token"`
	RefreshTokenSupabase string `json:"refresh_token_supabase"`
	ProjectID            string `json:"project_id"`
}

func (db DB) CreateRefreshToken(refreshToken string, refreshTokenSupabase string, projectID string) error {
	err := db.sql.Exec(`
		INSERT INTO refresh_tokens (refresh_token, refresh_token_supabase, project_id)
		VALUES (@refresh_token, @refresh_token_supabase, @project_id)
	`, sql.Named("refresh_token", refreshToken), sql.Named("refresh_token_supabase", refreshTokenSupabase), sql.Named("project_id", projectID)).Error
	if err != nil {
		return err
	}

	return nil
}

func (db DB) GetAndDeleteRefreshToken(refreshToken string) (*RefreshToken, error) {
	var refreshTokenStruct []RefreshToken

	err := db.sql.Raw(`
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

func (db DB) GetDevEmail(projectID string) (string, error) {
	var devEmail []string

	err := db.sql.Raw(`SELECT get_dev_email_project_id(@project_id)`, sql.Named("project_id", projectID)).Scan(&devEmail).Error
	if err != nil {
		return "", err
	}

	if len(devEmail) == 0 {
		return "", errors.New("Invalid project id")
	}

	return devEmail[0], nil
}

func (db DB) GetUserIDFromProjectAuthID(
	project string,
	authID string,
) (*string, error) {
	var results []ProjectUser

	err := db.sql.Find(&results, "auth_id = ? AND project_id = ?", authID, project).Error
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}

	return &results[0].ID, nil
}

func (db DB) CreateProjectUser(
	authID string,
	projectID string,
	monthlyCreditRateLimit *int,
) (*string, error) {
	var result *ProjectUser

	fmt.Println("Creating project user", authID, projectID, monthlyCreditRateLimit)
	err := db.sql.Create(&ProjectUserInsert{
		AuthID:                 authID,
		ProjectID:              projectID,
		MonthlyCreditRateLimit: monthlyCreditRateLimit,
	}).Scan(&result).Error
	if err != nil {
		return nil, err
	}

	return &result.ID, nil
}
