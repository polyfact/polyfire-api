package context

import (
	"errors"

	"github.com/polyfire/api/tokens"
)

type Priority int

const (
	HELPFUL   Priority = 1 // Can be reduced up to the minimum context size if needed. Can be dismissed in extreme cases. ex. Memory, web search, etc...
	IMPORTANT Priority = 2 // Should have the recommended context size if possible. ex. Chat history
	CRITICAL  Priority = 3 // Must always be present at the recommended size. ex. System prompts
)

// In order of importance
// Critical Recommended/Minimum > Important Minimum > Helpful Minimum > Important Recommended > Helpful Minimum

type ContentElement interface {
	GetMinimumContextSize() int
	GetRecommendedContextSize() int
	GetPriority() Priority
	GetContentFittingIn(token_count int) string
}

var CriticalDoesNotFitWarning = errors.New("Critical content does not fit in the context")

type contextElement struct {
	Minimum         string
	MinimumSize     int
	Recommended     string
	RecommendedSize int
	UseRecommended  bool
}

func GetContext(content []ContentElement, tokenLimit int) (string, error) {
	context := ""
	tokenCount := 0

	// First we get the critical elements directly in the context
	for _, item := range content {
		if item.GetPriority() == CRITICAL {
			added := item.GetContentFittingIn(tokenLimit)
			addedTokens := tokens.CountTokens("gpt-3.5-turbo", added)
			if addedTokens+tokenCount > tokenLimit {
				return context, CriticalDoesNotFitWarning
			}

			context += added
			tokenCount += addedTokens
		}
	}

	// Then we get the important elements with the minimum size
	importantContent := []contextElement{}
	for _, item := range content {
		if item.GetPriority() == IMPORTANT {
			minimumSize := item.GetMinimumContextSize()
			if (tokenCount + minimumSize) > tokenLimit {
				continue
			}

			importantContent = append(importantContent, contextElement{
				Minimum:         item.GetContentFittingIn(item.GetMinimumContextSize()),
				MinimumSize:     item.GetMinimumContextSize(),
				Recommended:     item.GetContentFittingIn(item.GetRecommendedContextSize()),
				RecommendedSize: item.GetRecommendedContextSize(),
				UseRecommended:  false,
			})

			tokenCount += minimumSize
		}
	}

	// Then we get the helpful elements with the minimum size
	helpfulContent := []contextElement{}
	for _, item := range content {
		if item.GetPriority() == HELPFUL {
			minimumSize := item.GetMinimumContextSize()
			if (tokenCount + minimumSize) > tokenLimit {
				continue
			}

			helpfulContent = append(helpfulContent, contextElement{
				Minimum:         item.GetContentFittingIn(item.GetMinimumContextSize()),
				MinimumSize:     item.GetMinimumContextSize(),
				Recommended:     item.GetContentFittingIn(item.GetRecommendedContextSize()),
				RecommendedSize: item.GetRecommendedContextSize(),
				UseRecommended:  false,
			})

			tokenCount += minimumSize
		}
	}

	// We try to fit the important elements with the recommended size
	for i, item := range importantContent {
		if (tokenCount + item.RecommendedSize - item.MinimumSize) > tokenLimit {
			continue
		}

		importantContent[i].UseRecommended = true
		tokenCount = tokenCount - item.MinimumSize + item.RecommendedSize
	}

	// We try to fit the helpful elements with the recommended size
	for i, item := range helpfulContent {
		if (tokenCount + item.RecommendedSize - item.MinimumSize) > tokenLimit {
			continue
		}

		helpfulContent[i].UseRecommended = true
		tokenCount = tokenCount - item.MinimumSize + item.RecommendedSize
	}

	// We add everything to the context
	for _, item := range importantContent {
		if item.UseRecommended {
			context += item.Recommended
		} else {
			context += item.Minimum
		}
	}

	for _, item := range helpfulContent {
		if item.UseRecommended {
			context += item.Recommended
		} else {
			context += item.Minimum
		}
	}

	return context, nil
}
