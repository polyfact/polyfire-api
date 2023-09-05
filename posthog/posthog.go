package posthog

import (
	"os"

	"github.com/posthog/posthog-go"
)

var POSTHOG_API_KEY = os.Getenv("POSTHOG_API_KEY")

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
	if POSTHOG_API_KEY == "" {
		return
	}

	client := posthog.New(POSTHOG_API_KEY)
	defer client.Close()

	properties := posthog.NewProperties()

	for key, value := range props {
		properties = properties.Set(key, value)
	}

	client.Enqueue(posthog.Capture{
		DistinctId: distinctId,
		Event:      eventName,
		Properties: properties,
	})
}
