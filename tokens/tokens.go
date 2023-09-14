package tokens

import (
	tiktoken "github.com/pkoukk/tiktoken-go"
	tiktoken_loader "github.com/pkoukk/tiktoken-go-loader"
)

func CountTokens(_model string, text string) (int) {
	encoding := "cl100k_base"

	tiktoken.SetBpeLoader(tiktoken_loader.NewOfflineLoader())
	tke, err := tiktoken.GetEncoding(encoding)
	if err != nil {
		return 0
	}

	token := tke.Encode(text, nil, nil)

	return len(token)
}
