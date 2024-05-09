package tokens

import (
	"errors"

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

func CountTokens(text string) int {
	token := tke.Encode(text, nil, nil)

	return len(token)
}

func SplitText(text string, chunkSize int) []string {
	splits := make([]string, 0)
	inputIDs := tke.Encode(text, nil, nil)

	startIdx := 0
	curIdx := len(inputIDs)
	if startIdx+chunkSize < curIdx {
		curIdx = startIdx + chunkSize
	}
	for startIdx < len(inputIDs) {
		chunkIDs := inputIDs[startIdx:curIdx]
		splits = append(splits, tke.Decode(chunkIDs))
		startIdx += chunkSize
		curIdx = startIdx + chunkSize
		if curIdx > len(inputIDs) {
			curIdx = len(inputIDs)
		}
	}
	return splits
}

func BatchText(input []string, maxBatchTokenSize int) ([][]string, error) {
	res := make([][]string, 0)
	currentBatch := make([]string, 0)
	currentBatchToken := 0

	for _, curr := range input {
		currTokens := CountTokens(curr)

		if currTokens > maxBatchTokenSize {
			return nil, errors.New("BatchText: One of the input is bigger than maxBatchTokenSize")
		}

		if currentBatchToken+currTokens > maxBatchTokenSize {
			res = append(res, currentBatch)
			currentBatchToken = 0
			currentBatch = make([]string, 0)
		}

		currentBatch = append(currentBatch, curr)
		currentBatchToken += currTokens
	}

	res = append(res, currentBatch)
	return res, nil
}
