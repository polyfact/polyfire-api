package posthog

import (
	"os"

	"github.com/posthog/posthog-go"
)

var PosthogAPIKey = os.Getenv("POSTHOG_API_KEY")

func IdentifyUser(authID string, userID string, email string) {
	if PosthogAPIKey == "" {
		return
	}

	client := posthog.New(PosthogAPIKey)
	defer client.Close()

	_ = client.Enqueue(posthog.Identify{
		DistinctId: authID,
		Properties: posthog.NewProperties().Set("email", email),
	})

	_ = client.Enqueue(posthog.Alias{
		DistinctId: authID,
		Alias:      userID,
	})
}

func Event(eventName string, distinctID string, props map[string]string) {
	if PosthogAPIKey == "" {
		return
	}

	client := posthog.New(PosthogAPIKey)
	defer client.Close()

	properties := posthog.NewProperties()

	for key, value := range props {
		properties = properties.Set(key, value)
	}

	_ = client.Enqueue(posthog.Capture{
		DistinctId: distinctID,
		Event:      eventName,
		Properties: properties,
	})
}
