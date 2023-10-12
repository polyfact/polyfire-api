package db

import (
	"encoding/json"
	"fmt"
	"regexp"
)

type ProjectUser struct {
	ID                     string `json:"id"`
	AuthID                 string `json:"auth_id"`
	ProjectID              string `json:"project_id"`
	MonthlyTokenRateLimit  *int   `json:"monthly_token_rate_limit"` // Deprecated
	MonthlyCreditRateLimit *int   `json:"monthly_credit_rate_limit"`
}

func (ProjectUser) TableName() string {
	return "project_users"
}

type ProjectUserInsert struct {
	AuthID                 string `json:"auth_id"`
	ProjectID              string `json:"project_id"`
	MonthlyTokenRateLimit  *int   `json:"monthly_token_rate_limit"` // Deprecated
	MonthlyCreditRateLimit *int   `json:"monthly_credit_rate_limit"`
}

func (ProjectUserInsert) TableName() string {
	return "project_users"
}

type Project struct {
	ID                            string      `json:"id"`
	Name                          string      `json:"name"`
	AuthID                        string      `json:"auth_id"`
	FreeUserInit                  bool        `json:"free_user_init"`
	DefaultMonthlyTokenRateLimit  *int        `json:"default_monthly_token_rate_limit"` // Deprecated
	DefaultMonthlyCreditRateLimit *int        `json:"default_monthly_credit_rate_limit"`
	FirebaseProjectID             string      `json:"firebase_project_id"`
	CustomAuthPublicKey           string      `json:"custom_auth_public_key"`
	AllowAnonymousAuth            bool        `json:"allow_anonymous_auth"`
	AuthorizedDomains             StringArray `json:"authorized_domains"`
}

func (Project) TableName() string {
	return "projects"
}

func GetProjectByID(id string) (*Project, error) {
	project := &Project{}

	matchUUID, _ := regexp.MatchString(UUIDRegexp, id)
	matchSlug, _ := regexp.MatchString(SlugRegexp, id)

	if !matchUUID && !matchSlug {
		return nil, fmt.Errorf("Invalid identifier")
	}

	var err error

	if matchUUID {
		err = DB.First(project, "id = ?", id).Error
	} else {
		err = DB.First(project, "slug = ?", id).Error
	}

	if err != nil {
		return nil, err
	}

	return project, nil
}

func GetProjectUserByID(id string) (*ProjectUser, error) {
	var results []ProjectUser

	DB.Find(&results, "id = ?", id)

	if len(results) == 0 {
		return nil, nil
	}

	return &results[0], nil
}

func GetProjectForUserId(user_id string) (*string, error) {
	var results []ProjectUser

	DB.Find(&results, "id = ?", user_id)

	if len(results) == 0 {
		return nil, nil
	}

	return &results[0].ProjectID, nil
}

type AuthUser struct {
	Usage       int    `json:"usage"`
	RateLimit   int    `json:"rate_limit,omitempty"`
	Premium     bool   `json:"premium"`
	OpenAIToken string `json:"openai_token,omitempty"`
	OpenAIOrg   string `json:"openai_org,omitempty"`
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
