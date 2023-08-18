package image_generation

import (
	"context"
	"os"
	"time"

	openai "github.com/rakyll/openai-go"
	image "github.com/rakyll/openai-go/image"
)

func Generate(prompt string) (*image.CreateResponse, error) {
	session := openai.NewSession(os.Getenv("OPENAI_API_KEY"))

	(*((*session).HTTPClient)).Timeout = 600 * time.Second
	(*session).OrganizationID = os.Getenv("OPENAI_ORGANIZATION")

	client := image.NewClient(session)

	params := image.CreateParams{
		Prompt: prompt,
	}

	ctx := context.Background()

	response, err := client.Create(ctx, &params)
	if err != nil {
		return nil, err
	}

	return response, nil
}
