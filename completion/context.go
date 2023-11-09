package completion

import (
	"context"
	"sync"

	completionContext "github.com/polyfire/api/completion/context"
	options "github.com/polyfire/api/llm/providers/options"
)

func parseMemoryIDArray(memoryID interface{}) []string {
	var memoryIDArray []string

	var str string
	var ok bool
	var array []interface{}
	if str, ok = memoryID.(string); ok {
		memoryIDArray = append(memoryIDArray, str)
	} else if array, ok = memoryID.([]interface{}); ok {
		for _, item := range array {
			if str, ok = item.(string); ok {
				memoryIDArray = append(memoryIDArray, str)
			}
		}
	}
	return memoryIDArray
}

func launchContextFillingGoRouting(
	wg *sync.WaitGroup,
	contextElements *[]completionContext.ContentElement,
	callback func() (completionContext.ContentElement, error),
) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		ce, err := callback()
		if err != nil || ce == nil {
			return
		}

		*contextElements = append(*contextElements, ce)
	}()
}

const MaxContentLength = 4000

func GetContextString(
	ctx context.Context,
	userID string,
	input GenerateRequestBody,
	callback options.ProviderCallback,
	opts *options.ProviderOptions,
) (string, []string, error) {
	var wg sync.WaitGroup
	contextElements := make([]completionContext.ContentElement, 0)

	launchContextFillingGoRouting(&wg, &contextElements, func() (completionContext.ContentElement, error) {
		return completionContext.GetMemory(ctx, userID, parseMemoryIDArray(input.MemoryID), input.Task)
	})

	var warnings []string
	launchContextFillingGoRouting(&wg, &contextElements, func() (completionContext.ContentElement, error) {
		var systemPrompt completionContext.ContentElement
		var err error
		systemPrompt, warnings, err = completionContext.GetSystemPrompt(
			userID,
			input.SystemPromptID,
			input.SystemPrompt,
			input.ChatID,
		)
		return systemPrompt, err
	})

	if input.WebRequest {
		launchContextFillingGoRouting(&wg, &contextElements, func() (completionContext.ContentElement, error) {
			return completionContext.GetWebContext(input.Task)
		})
	}

	if input.ChatID != nil && len(*input.ChatID) > 0 {
		err := AddToChatHistory(userID, input.Task, *input.ChatID, callback, opts)
		if err != nil {
			return "", warnings, err
		}

		launchContextFillingGoRouting(&wg, &contextElements, func() (completionContext.ContentElement, error) {
			return completionContext.GetChatHistoryContext(userID, *input.ChatID)
		})
	}

	wg.Wait()

	contextString, err := completionContext.GetContext(contextElements, MaxContentLength)
	if err != nil {
		return "", warnings, err
	}

	return contextString, warnings, nil
}
