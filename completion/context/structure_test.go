package context

import (
	"strings"
	"testing"

	"github.com/polyfire/api/tokens"
	"github.com/polyfire/api/utils"
)

type TestContentElement1 struct{}

func (TestContentElement1) GetMinimumContextSize() int {
	return 1
}

func (TestContentElement1) GetRecommendedContextSize() int {
	return 5
}

func (TestContentElement1) GetPriority() Priority {
	return HELPFUL
}

func (TestContentElement1) GetOrderIndex() int {
	return 1
}

func (TestContentElement1) GetContentFittingIn(i int) string {
	return strings.Repeat("abc", i)
}

type TestContentElement2 struct {
	TestContentElement1
}

func (TestContentElement2) GetPriority() Priority {
	return IMPORTANT
}

func (TestContentElement2) GetContentFittingIn(i int) string {
	return strings.Repeat("def", i)
}

func TestContextStructureWithLotOfSpace(t *testing.T) {
	utils.SetLogLevel("WARN")

	maxTokens := 20

	contextElements := []ContentElement{TestContentElement1{}, TestContentElement2{}}

	result, err := GetContext(contextElements, maxTokens)
	if err != nil {
		t.Fatalf(`GetContext returned an error : %v`, err)
	}

	if tokens.CountTokens(result) != maxTokens {
		t.Fatalf(
			`Result string is different than max Tokens. %v > Expected size (%v)`,
			tokens.CountTokens(result),
			maxTokens,
		)
	}

	if !strings.Contains(result, "abcabcabcabcabc") {
		t.Fatalf(
			`HELPFUL content should be at recommended size when total recommended size < maxTokens`,
		)
	}

	if !strings.Contains(result, "defdefdefdefdefdefdefdefdefdefdefdefdefdefdef") {
		t.Fatalf(`IMPORTANT content should fill up all available space`)
	}
}

func TestContextStructureWithSomeSpace(t *testing.T) {
	utils.SetLogLevel("WARN")

	maxTokens := 6

	contextElements := []ContentElement{TestContentElement1{}, TestContentElement2{}}

	result, err := GetContext(contextElements, maxTokens)
	if err != nil {
		t.Fatalf(`GetContext returned an error : %v`, err)
	}

	if tokens.CountTokens(result) != maxTokens {
		t.Fatalf(
			`Result string is different than max Tokens. %v > Expected size (%v)`,
			tokens.CountTokens(result),
			maxTokens,
		)
	}

	if !strings.Contains(result, "abc") {
		t.Fatalf(
			`HELPFUL content should be at minimum size when total recommended size > maxTokens`,
		)
	}

	if !strings.Contains(result, "defdefdefdefdef") {
		t.Fatalf(`IMPORTANT content should be at recommended size when possible`)
	}
}

func TestContextStructureWithFewSpace(t *testing.T) {
	utils.SetLogLevel("WARN")

	maxTokens := 2

	contextElements := []ContentElement{TestContentElement1{}, TestContentElement2{}}

	result, err := GetContext(contextElements, maxTokens)
	if err != nil {
		t.Fatalf(`GetContext returned an error : %v`, err)
	}

	if tokens.CountTokens(result) != maxTokens {
		t.Fatalf(
			`Result string is different than max Tokens. %v > Expected size (%v)`,
			tokens.CountTokens(result),
			maxTokens,
		)
	}

	if !strings.Contains(result, "abc") {
		t.Fatalf(`HELPFUL content should be at least at minimum size`)
	}

	if !strings.Contains(result, "def") {
		t.Fatalf(`IMPORTANT content should be at least at minimum size`)
	}
}

func TestContextStructureWithVeryFewSpace(t *testing.T) {
	utils.SetLogLevel("WARN")

	maxTokens := 1

	contextElements := []ContentElement{TestContentElement1{}, TestContentElement2{}}

	result, err := GetContext(contextElements, maxTokens)
	if err != nil {
		t.Fatalf(`GetContext returned an error : %v`, err)
	}

	if tokens.CountTokens(result) != maxTokens {
		t.Fatalf(
			`Result string is different than max Tokens. %v > Expected size (%v)`,
			tokens.CountTokens(result),
			maxTokens,
		)
	}

	if !strings.Contains(result, "def") {
		t.Fatalf(`IMPORTANT content should be at least at minimum size`)
	}
}
