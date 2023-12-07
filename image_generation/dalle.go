package imagegeneration

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"github.com/polyfire/api/utils"
	openai "github.com/sashabaranov/go-openai"
)

func DALLEGenerate(ctx context.Context, prompt string, model string) (io.Reader, error) {
	var config openai.ClientConfig
	customToken, ok := ctx.Value(utils.ContextKeyOpenAIToken).(string)
	if ok {
		config = openai.DefaultConfig(customToken)
		customOrg, ok := ctx.Value(utils.ContextKeyOpenAIOrg).(string)
		if ok {
			config.OrgID = customOrg
		}
	} else {
		config = openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
		config.OrgID = os.Getenv("OPENAI_ORGANIZATION")
	}
	client := openai.NewClientWithConfig(config)

	req := openai.ImageRequest{
		Prompt:         prompt,
		Model:          ModelToOpenAIFormat(model),
		Size:           openai.CreateImageSize1024x1024,
		ResponseFormat: openai.CreateImageResponseFormatB64JSON,
		N:              1,
	}

	respBase64, err := client.CreateImage(ctx, req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	imgBytes, err := base64.StdEncoding.DecodeString(respBase64.Data[0].B64JSON)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	r := bytes.NewReader(imgBytes)

	return r, nil
}
