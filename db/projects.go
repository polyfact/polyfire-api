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

func GetProjectForUserID(userID string) (*string, error) {
	var results []ProjectUser

	DB.Find(&results, "id = ?", userID)

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

func GetDevAuthUserForUserIDProject(userID string) (AuthUser, error) {
	client, err := CreateClient()
	if err != nil {
		return AuthUser{}, err
	}

	params := AuthUserProject{
		UserID: userID,
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

func ProjectReachedRateLimit(userID string) (bool, error) {
	usageRateLimit, err := GetDevAuthUserForUserIDProject(userID)
	if err != nil {
		return false, err
	}

	return usageRateLimit.Usage >= usageRateLimit.RateLimit, nil
}

func ProjectIsPremium(userID string) (bool, error) {
	authUsers, err := GetDevAuthUserForUserIDProject(userID)
	if err != nil {
		return false, err
	}

	return authUsers.Premium, nil
}

type Model struct {
	ID       int    `json:"id"`
	Model    string `json:"model"`
	Provider string `json:"provider"`
}

func (Model) TableName() string {
	return "models"
}

func GetModelByAliasAndProjectID(alias string, projectID string, modelType string) (*Model, error) {
	model := &Model{}

	err := DB.Raw(
		"SELECT models.* FROM models JOIN model_aliases ON model_aliases.model_id = models.id WHERE model_aliases.alias = ? AND model_aliases.project_id = ? AND models.type = ?",
		alias,
		projectID,
		modelType,
	).Scan(model).Error
	if err != nil {
		return nil, err
	}

	return model, nil
}
