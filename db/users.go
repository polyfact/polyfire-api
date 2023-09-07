package db

type UserAuth struct {
	Id      string `json:"id"`
	Version int    `json:"version,omitempty"`
}

func GetVersionForUser(user_id string) (int, error) {
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
