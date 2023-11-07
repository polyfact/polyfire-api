package imagegeneration

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/polyfire/api/utils"
	openai "github.com/rakyll/openai-go"
	image "github.com/rakyll/openai-go/image"
)

func DALLEGenerate(ctx context.Context, prompt string) (io.Reader, error) {
	var session *openai.Session
	customToken, ok := ctx.Value(utils.ContextKeyOpenAIToken).(string)
	if ok {
		session = openai.NewSession(customToken)
		customOrg, ok := ctx.Value(utils.ContextKeyOpenAIOrg).(string)
		if ok {
			(*session).OrganizationID = customOrg
		}
	} else {
		session = openai.NewSession(os.Getenv("OPENAI_API_KEY"))
		(*session).OrganizationID = os.Getenv("OPENAI_ORGANIZATION")
	}

	(*((*session).HTTPClient)).Timeout = 600 * time.Second

	client := image.NewClient(session)

	params := image.CreateParams{
		Prompt: prompt,
	}

	imageGenerationCtx := context.Background()

	response, err := client.Create(imageGenerationCtx, &params)
	if err != nil {
		return nil, err
	}

	return (response.Data[0]).Reader()
}
