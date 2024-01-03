package imagegeneration

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"

	providers "github.com/polyfire/api/llm/providers"
	openai "github.com/sashabaranov/go-openai"
)

func DALLEGenerate(ctx context.Context, prompt string, model string) (io.Reader, error) {
	client := providers.NewOpenAIStreamProvider(ctx, model).Client

	req := openai.ImageRequest{
		Prompt:         prompt,
		Model:          model,
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
