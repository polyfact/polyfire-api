package completion

import (
	"testing"

	options "github.com/polyfire/api/llm/providers/options"
)

func TestAddSpace(t *testing.T) {
	prompt := "My name is"
	generation := make(chan options.Result, 5)

	result := AddSpaceIfNeeded(prompt, generation)

	generation <- options.Result{Result: "john"}
	generation <- options.Result{Result: " doe"}

	close(generation)

	resultStr := ""

	for v := range result {
		resultStr += v.Result
	}

	if resultStr != " john doe" {
		t.Fatalf(
			`AddSpaceIfNeeded("My name is", "john doe") should give " john doe". Result = "%v"`,
			resultStr,
		)
	}
}

func TestDontAddSpace(t *testing.T) {
	prompt := "My name i"
	generation := make(chan options.Result, 5)

	result := AddSpaceIfNeeded(prompt, generation)

	generation <- options.Result{Result: "s"}
	generation <- options.Result{Result: " john"}
	generation <- options.Result{Result: " doe"}

	close(generation)

	resultStr := ""

	for v := range result {
		resultStr += v.Result
	}

	if resultStr != "s john doe" {
		t.Fatalf(
			`AddSpaceIfNeeded("My name i", "s john doe") should give "s john doe". Result = "%v"`,
			resultStr,
		)
	}
}
