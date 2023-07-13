package split

import (
	"math"
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
