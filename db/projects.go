package db

type Project struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	AuthID       string `json:"auth_id"`
	FreeUserInit bool   `json:"free_user_init"`
}

type ProjectUser struct {
	ID                    string `json:"id"`
	AuthID                string `json:"auth_id"`
	ProjectID             string `json:"project_id"`
	MonthlyTokenRateLimit *int   `json:"monthly_token_rate_limit"`
}

type ProjectUserInsert struct {
	AuthID    string `json:"auth_id"`
	ProjectID string `json:"project_id"`
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

	if projectUser == nil || projectUser.MonthlyTokenRateLimit == nil {
		return false, nil
	}

	tokenUsage := GetUserIdMonthlyTokenUsage(id)

	return tokenUsage >= *projectUser.MonthlyTokenRateLimit, nil
}
