package tokens

import (
	tiktoken "github.com/pkoukk/tiktoken-go"
	tiktoken_loader "github.com/pkoukk/tiktoken-go-loader"
)

func initEncoding() *tiktoken.Tiktoken {
	encoding := "cl100k_base"
	tiktoken.SetBpeLoader(tiktoken_loader.NewOfflineLoader())
	tke, err := tiktoken.GetEncoding(encoding)
	if err != nil {
		panic(err)
	}

	return tke
}

var tke = initEncoding()

func CountTokens(_model string, text string) int {
	token := tke.Encode(text, nil, nil)

	return len(token)
}
