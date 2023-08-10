package posthog

import (
	"fmt"
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
	fmt.Printf("[Event] Identify { \"distinctId\": \"%v\", \"email\": \"%v\" }\n", auth_id, email)
	fmt.Printf("[Event] Alias { \"distinctId\": \"%v\", \"alias\": \"%v\" }\n", auth_id, user_id)
	if POSTHOG_API_KEY == "" {
		return
	}

	client := posthog.New(POSTHOG_API_KEY)
	defer client.Close()

	err := client.Enqueue(posthog.Identify{
		DistinctId: auth_id,
		Properties: posthog.NewProperties().Set("email", email),
	})

	fmt.Printf("%v\n", err)

	client.Enqueue(posthog.Alias{
		DistinctId: auth_id,
		Alias:      user_id,
	})
}

func GenerateEvent(distinctId string, model string, tokenUsageInput int, tokenUsageOutput int) {
	pId, _ := getProjectForUserId(distinctId)
	projectId := "00000000-0000-0000-0000-000000000000"

	if pId != nil {
		projectId = *pId
	}

	fmt.Printf(
		"[Event] Event Generation { \"distinctId\": \"%v\", \"projectId\": \"%v\", \"model\": \"%v\", \"tokenUsageInput\": %v, \"tokenUsageOutput\": %v }\n",
		distinctId,
		projectId,
		model,
		tokenUsageInput,
		tokenUsageOutput,
	)
	if POSTHOG_API_KEY == "" {
		return
	}

	client := posthog.New(POSTHOG_API_KEY)
	defer client.Close()

	err := client.Enqueue(posthog.Capture{
		DistinctId: distinctId,
		Event:      "Generation",
		Properties: posthog.NewProperties().
			Set("projectId", projectId).
			Set("model", model).
			Set("tokenUsageInput", tokenUsageInput).
			Set("tokenUsageOutput", tokenUsageOutput),
	})
}
