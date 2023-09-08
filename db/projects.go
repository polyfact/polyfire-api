package db

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
	client, err := CreateClient()
	if err != nil {
		return nil, err
	}

	var result Project

	_, err = client.From("projects").
		Select("*", "exact", false).
		Eq("id", id).
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

	// TODO: Remove the tokenUsage check once we've migrated to the new rate limit system
	tokenUsage, err := GetUserIdMonthlyTokenUsage(id)
	if err != nil {
		return false, err
	}

	if tokenUsage >= *projectUser.MonthlyTokenRateLimit {
		return true, nil
	}

	creditUsage, err := GetUserIdMonthlyCreditUsage(id)
	if err != nil {
		return false, err
	}

	if creditUsage >= *projectUser.MonthlyCreditRateLimit {
		return true, nil
	}
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
