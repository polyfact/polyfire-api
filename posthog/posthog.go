package posthog

import (
	"os"

	db "github.com/polyfact/api/db"
	"github.com/posthog/posthog-go"
)

var POSTHOG_API_KEY = os.Getenv("POSTHOG_API_KEY")

type ProjectUser struct {
	ID        string `json:"id"`
	AuthID    string `json:"auth_id"`
	ProjectID string `json:"project_id"`
}

func getProjectForUserId(user_id string) (*string, error) {
	client, err := db.CreateClient()
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

func IdentifyUser(auth_id string, user_id string, email string) {
	if POSTHOG_API_KEY == "" {
		return
	}

	client := posthog.New(POSTHOG_API_KEY)
	defer client.Close()

	client.Enqueue(posthog.Identify{
		DistinctId: auth_id,
		Properties: posthog.NewProperties().Set("email", email),
	})

	client.Enqueue(posthog.Alias{
		DistinctId: auth_id,
		Alias:      user_id,
	})
}

func Event(eventName string, distinctId string, props map[string]string) {
	pId, _ := getProjectForUserId(distinctId)
	projectId := "00000000-0000-0000-0000-000000000000"

	if pId != nil {
		projectId = *pId
	}

	if POSTHOG_API_KEY == "" {
		return
	}

	client := posthog.New(POSTHOG_API_KEY)
	defer client.Close()

	properties := posthog.NewProperties().Set("projectId", projectId)

	for key, value := range props {
		properties = properties.Set(key, value)
	}

	client.Enqueue(posthog.Capture{
		DistinctId: distinctId,
		Event:      eventName,
		Properties: properties,
	})
}
