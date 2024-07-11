package context

import (
	"errors"
	"sort"

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
	GetOrderIndex() int
	GetContentFittingIn(tokenCount int) string
}

var ErrCriticalDoesNotFit = errors.New("Critical content does not fit in the context")

type contextElement struct {
	ContentElement  ContentElement
	Minimum         string
	MinimumSize     int
	Recommended     string
	RecommendedSize int
	UseRecommended  bool
	OrderIndex      int
}

type contextElementList []contextElement

func (cel *contextElementList) Len() int {
	return len(*cel)
}

func (cel *contextElementList) Less(i, j int) bool {
	return (*cel)[i].OrderIndex < (*cel)[j].OrderIndex
}

func (cel *contextElementList) Swap(i, j int) {
	(*cel)[i], (*cel)[j] = (*cel)[j], (*cel)[i]
}

func contextElementFromContentElement(content ContentElement) contextElement {
	return contextElement{
		ContentElement:  content,
		Minimum:         content.GetContentFittingIn(content.GetMinimumContextSize()),
		MinimumSize:     content.GetMinimumContextSize(),
		Recommended:     content.GetContentFittingIn(content.GetRecommendedContextSize()),
		RecommendedSize: content.GetRecommendedContextSize(),
		UseRecommended:  false,
		OrderIndex:      content.GetOrderIndex(),
	}
}

func GetContext(content []ContentElement, tokenLimit int) (string, error) {
	tokenCount := 0

	criticalContent := []contextElement{}
	// First we get the critical elements directly in the context
	for _, item := range content {
		if item.GetPriority() == CRITICAL {
			added := item.GetContentFittingIn(tokenLimit)
			addedTokens := tokens.CountTokens(added)
			if addedTokens+tokenCount > tokenLimit {
				return "", ErrCriticalDoesNotFit
			}

			criticalContent = append(criticalContent, contextElementFromContentElement(item))

			tokenCount += addedTokens
		}
	}

	// Then we get the important elements with the minimum size
	importantAndHelpfulContent := []contextElement{}
	for _, item := range content {
		if item.GetPriority() == IMPORTANT {
			minimumSize := item.GetMinimumContextSize()
			if (tokenCount + minimumSize) > tokenLimit {
				continue
			}

			importantAndHelpfulContent = append(
				importantAndHelpfulContent,
				contextElementFromContentElement(item),
			)

			tokenCount += minimumSize
		}
	}

	// Then we get the helpful elements with the minimum size
	for _, item := range content {
		if item.GetPriority() == HELPFUL {
			minimumSize := item.GetMinimumContextSize()
			if (tokenCount + minimumSize) > tokenLimit {
				continue
			}

			importantAndHelpfulContent = append(
				importantAndHelpfulContent,
				contextElementFromContentElement(item),
			)

			tokenCount += minimumSize
		}
	}

	// We try to increase the size of the important and helpful elements (in the order of importance we added them in) to the recommended size
	for i, item := range importantAndHelpfulContent {
		importantAndHelpfulContent[i].Recommended = item.ContentElement.GetContentFittingIn(
			item.RecommendedSize,
		)
		importantAndHelpfulContent[i].RecommendedSize = tokens.CountTokens(
			importantAndHelpfulContent[i].Recommended,
		)

		if (tokenCount + importantAndHelpfulContent[i].RecommendedSize - importantAndHelpfulContent[i].MinimumSize) > tokenLimit {
			continue
		}

		importantAndHelpfulContent[i].UseRecommended = true
		tokenCount = tokenCount - importantAndHelpfulContent[i].MinimumSize + importantAndHelpfulContent[i].RecommendedSize
	}

	// We try to increase the size of the important and helpful elements (in the order of importance we added them in) as much as possible
	for i, item := range importantAndHelpfulContent {
		size := item.MinimumSize
		if item.UseRecommended {
			size = item.RecommendedSize
		}
		recommended := item.ContentElement.GetContentFittingIn(
			tokenLimit - (tokenCount - size),
		)
		recommendedSize := tokens.CountTokens(recommended)

		if (tokenCount + recommendedSize - size) > tokenLimit {
			continue
		}
		importantAndHelpfulContent[i].Recommended = recommended
		importantAndHelpfulContent[i].RecommendedSize = recommendedSize
		importantAndHelpfulContent[i].UseRecommended = true
		tokenCount = tokenCount - size + importantAndHelpfulContent[i].RecommendedSize
	}

	// We merge all the context and sort them by the order we want to get them in the context prompt
	var context contextElementList = append(criticalContent, importantAndHelpfulContent...)

	sort.Sort(&context)

	result := ""
	for _, item := range context {
		if item.UseRecommended {
			result += item.Recommended
		} else {
			result += item.Minimum
		}
	}

	return result, nil
}
