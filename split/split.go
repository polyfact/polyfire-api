package split

import (
	"fmt"
	"math"

	llm "github.com/polyfact/api/llm"
)

func stirling(n float64) float64 {
	return math.Pow((n/math.E), n) * math.Sqrt(2.0*math.Pi*n)
}

func binomial_score(curr int, max int) float64 {
	n := 30.0
	k := (n-2.0)*(float64(curr)/float64(max)) + 1
	return math.Sqrt((stirling(n) / (stirling(k) * stirling(n-k))) * math.Pow(0.5, float64(k)) * math.Pow(0.5, float64(n-k)))
}

func new_line_score(s string, i int) float64 {
	if len(s) == i+1 || s[i] != '\n' {
		return 1.0
	}

	if s[i+1] == '\n' {
		return 50.0
	}
	return 5.0
}

func Split(s string) (string, string) {
	max_score := 0.0
	max_score_i := 0
	last_new_line := -1
	for i := 0; i < len(s); i++ {
		score := binomial_score(i, len(s)) * new_line_score(s, i)

		if s[i] == '\n' {
			if s[last_new_line+1] != '\t' {
				score *= 50.0
			}
			last_new_line = i
		}

		if s[i] == ' ' {
			score *= 2.0
		}

		if score > max_score {
			max_score = score
			max_score_i = i
		}
	}

	return s[:max_score_i], s[max_score_i:]
}

func BinarySplit(s string, max_token int) []string {
	if TokenCount(s) < max_token {
		return []string{s}
	}

	left, right := Split(s)

	return append(BinarySplit(left, max_token), BinarySplit(right, max_token)...)
}

func GenerateWithSplit(type_format any, s string, c *func(string, int, int), max_token int) (llm.Result, error) {
	b_split := BinarySplit(s, max_token)

	task := ExtractTaskFromSplits(b_split)

	fmt.Printf("task %v", task)

	var merged_results interface{}
	token_usage := llm.TokenUsage{Input: 0, Output: 0}
	for i := 0; i < len(b_split); i++ {
		b_split[i] = task + "\n\n" + b_split[i]

		result, err := llm.Generate(type_format, b_split[i], c)

		token_usage.Input += result.TokenUsage.Input
		token_usage.Output += result.TokenUsage.Output

		fmt.Printf("Chunk %v/%v\n", i, len(b_split))
		fmt.Printf("Token Usage %v\n", token_usage)

		if err != nil {
			return result, err
		}

		if i == 0 {
			merged_results = result.Result
		} else {
			merged_results = Merge(merged_results, result.Result)
		}
	}

	return llm.Result{
		Result:     merged_results,
		TokenUsage: token_usage,
	}, nil
}
