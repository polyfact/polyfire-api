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

func GenerateEvent(distinctId string, model string, tokenUsageInput int, tokenUsageOutput int) {
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

	client.Enqueue(posthog.Capture{
		DistinctId: distinctId,
		Event:      "Generation",
		Properties: posthog.NewProperties().
			Set("projectId", projectId).
			Set("model", model).
			Set("tokenUsageInput", tokenUsageInput).
			Set("tokenUsageOutput", tokenUsageOutput),
	})
}

func SimpleEvent(eventName string) func(distinctId string) {
	return func(distinctId string) {
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

		client.Enqueue(posthog.Capture{
			DistinctId: distinctId,
			Event:      eventName,
			Properties: posthog.NewProperties().
				Set("projectId", projectId),
		})
	}
}

var (
	CreateChatEvent      = SimpleEvent("CreateChat")
	GetChatHistoryEvent  = SimpleEvent("GetChatHistory")
	TranscribeEvent      = SimpleEvent("Transcribe")
	ImageGenerationEvent = SimpleEvent("ImageGeneration")
	GetMemoryEvent       = SimpleEvent("GetMemory")
	AddToMemoryEvent     = SimpleEvent("AddToMemory")
	CreateMemoryEvent    = SimpleEvent("CreateMemory")

	GetKVEvent           = SimpleEvent("GetKV")
	SetKVEvent           = SimpleEvent("SetKV")
	GetPromptByNameEvent = SimpleEvent("GetPromptByName")
	GetPromptByIdEvent   = SimpleEvent("GetPromptById")
	GetAllPromptsEvent   = SimpleEvent("GetAllPrompts")
	CreatePromptEvent    = SimpleEvent("CreatePrompt")
	UpdatePromptEvent    = SimpleEvent("UpdatePrompt")
	DeletePromptEvent    = SimpleEvent("DeletePrompt")
)
