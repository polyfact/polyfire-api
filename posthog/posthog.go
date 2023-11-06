package posthog

import (
	"os"

	"github.com/posthog/posthog-go"
)

var PosthogApiKey = os.Getenv("POSTHOG_API_KEY")

func IdentifyUser(auth_id string, user_id string, email string) {
	if PosthogApiKey == "" {
		return
	}

	client := posthog.New(PosthogApiKey)
	defer client.Close()

	_ = client.Enqueue(posthog.Identify{
		DistinctId: auth_id,
		Properties: posthog.NewProperties().Set("email", email),
	})

	_ = client.Enqueue(posthog.Alias{
		DistinctId: auth_id,
		Alias:      user_id,
	})
}

func Event(eventName string, distinctId string, props map[string]string) {
	if PosthogApiKey == "" {
		return
	}

	client := posthog.New(PosthogApiKey)
	defer client.Close()

	properties := posthog.NewProperties()

	for key, value := range props {
		properties = properties.Set(key, value)
	}

	_ = client.Enqueue(posthog.Capture{
		DistinctId: distinctId,
		Event:      eventName,
		Properties: properties,
	})
}
