package db

import (
	"encoding/json"
	"regexp"
	"fmt"
)

type ProjectUser struct {
	ID                     string `json:"id"`
	AuthID                 string `json:"auth_id"`
	ProjectID              string `json:"project_id"`
	MonthlyTokenRateLimit  *int   `json:"monthly_token_rate_limit"` // Deprecated
	MonthlyCreditRateLimit *int   `json:"monthly_credit_rate_limit"`
}

type ProjectUserInsert struct {
	AuthID                 string `json:"auth_id"`
	ProjectID              string `json:"project_id"`
	MonthlyTokenRateLimit  *int   `json:"monthly_token_rate_limit"` // Deprecated
	MonthlyCreditRateLimit *int   `json:"monthly_credit_rate_limit"`
}

type Project struct {
	ID                            string `json:"id"`
	Name                          string `json:"name"`
	AuthID                        string `json:"auth_id"`
	FreeUserInit                  bool   `json:"free_user_init"`
	DefaultMonthlyTokenRateLimit  *int   `json:"default_monthly_token_rate_limit"` // Deprecated
	DefaultMonthlyCreditRateLimit *int   `json:"default_monthly_credit_rate_limit"`
	FirebaseProjectID             string `json:"firebase_project_id"`
}

func GetProjectByID(id string) (*Project, error) {
	matchUUID, _ := regexp.MatchString("^[a-f0-9]{8}-[a-f0-9]{4}-[1-5][a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$", id)

	matchSlug, _ := regexp.MatchString("^[a-z0-9-_]*$", id)

	if !matchUUID && !matchSlug {
		return nil, fmt.Errorf("Invalid project id")
	}

	client, err := CreateClient()
	if err != nil {
		return nil, err
	}

	var result Project

	var id_field string
	if matchUUID {
		id_field = "id"
	} else {
		id_field = "slug"
	}

	_, err = client.From("projects").
		Select("*", "exact", false).
		Eq(id_field, id).
		Single().
		ExecuteTo(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func GetProjectUserByID(id string) (*ProjectUser, error) {
	client, err := CreateClient()
	if err != nil {
		return nil, err
	}

	var results []ProjectUser

	_, err = client.From("project_users").
		Select("*", "exact", false).
		Eq("id", id).
		ExecuteTo(&results)

	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}

	return &results[0], nil
}

func UserReachedRateLimit(id string) (bool, error) {
	projectUser, err := GetProjectUserByID(id)
	if err != nil {
		return false, err
	}

	if projectUser == nil || projectUser.MonthlyTokenRateLimit == nil || projectUser.MonthlyCreditRateLimit == nil {
		return false, nil
	}

	creditUsage, err := GetUserIdMonthlyCreditUsage(id)
	if err != nil {
		return false, err
	}

	if creditUsage >= *projectUser.MonthlyCreditRateLimit {
		return true, nil
	}

	return false, nil
}

func GetProjectForUserId(user_id string) (*string, error) {
	client, err := CreateClient()
	if err != nil {
		return nil, err
	}

	var results []ProjectUser

	_, err = client.From("project_users").
		Select("*", "exact", false).
		Eq("id", user_id).
		ExecuteTo(&results)

	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}

	return &results[0].ProjectID, nil
}

type AuthUser struct {
	Usage     int  `json:"usage"`
	RateLimit int  `json:"rate_limit,omitempty"`
	Premium   bool `json:"premium"`
}

type AuthUserProject struct {
	UserID string `json:"param_user_id"`
}

func GetDevAuthUserForUserIDProject(user_id string) (AuthUser, error) {
	client, err := CreateClient()
	if err != nil {
		return AuthUser{}, err
	}

	params := AuthUserProject{
		UserID: user_id,
	}

	response := client.Rpc("get_monthly_token_usage_user_id_projects", "", params)

	var result AuthUser
	err = json.Unmarshal([]byte(response), &result)
	if err != nil {
		return AuthUser{}, err
	}

	if result.RateLimit == 0 {
		result.RateLimit = 50_000_000
	}

	return result, nil
}

func ProjectReachedRateLimit(user_id string) (bool, error) {
	usage_rate_limit, err := GetDevAuthUserForUserIDProject(user_id)
	if err != nil {
		return false, err
	}

	return usage_rate_limit.Usage >= usage_rate_limit.RateLimit, nil
}

func ProjectIsPremium(user_id string) (bool, error) {
	auth_users, err := GetDevAuthUserForUserIDProject(user_id)
	if err != nil {
		return false, err
	}

	return auth_users.Premium, nil
}
