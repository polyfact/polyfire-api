package image_generation

import (
	"context"
	"io"
	"os"
	"time"

	openai "github.com/rakyll/openai-go"
	image "github.com/rakyll/openai-go/image"
)

func DALLEGenerate(prompt string) (io.Reader, error) {
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

	return (response.Data[0]).Reader()
}
