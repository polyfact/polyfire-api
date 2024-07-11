package completion

import (
	"context"
	"sync"

	completionContext "github.com/polyfire/api/completion/context"
	"github.com/polyfire/api/llm/providers/options"
	"github.com/polyfire/api/utils"
)

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

	launchContextFillingGoRouting(
		&wg,
		&contextElements,
		func() (completionContext.ContentElement, error) {
			return completionContext.GetMemory(
				ctx,
				userID,
				utils.StringOptionalArray(input.MemoryID),
				input.Task,
			)
		},
	)

	var warnings []string
	launchContextFillingGoRouting(
		&wg,
		&contextElements,
		func() (completionContext.ContentElement, error) {
			var systemPrompt completionContext.ContentElement
			var err error
			systemPrompt, warnings, err = completionContext.GetSystemPrompt(
				ctx,
				userID,
				input.SystemPromptID,
				input.SystemPrompt,
				input.ChatID,
			)
			return systemPrompt, err
		},
	)

	if input.WebRequest {
		launchContextFillingGoRouting(
			&wg,
			&contextElements,
			func() (completionContext.ContentElement, error) {
				return completionContext.GetWebContext(input.Task)
			},
		)
	}

	if input.ChatID != nil && len(*input.ChatID) > 0 {
		err := AddToChatHistory(ctx, userID, input.Task, *input.ChatID, callback, opts)
		if err != nil {
			return "", warnings, err
		}

		launchContextFillingGoRouting(
			&wg,
			&contextElements,
			func() (completionContext.ContentElement, error) {
				return completionContext.GetChatHistoryContext(ctx, userID, *input.ChatID)
			},
		)
	}

	wg.Wait()

	contextString, err := completionContext.GetContext(contextElements, MaxContentLength)
	if err != nil {
		return "", warnings, err
	}

	return contextString, warnings, nil
}
