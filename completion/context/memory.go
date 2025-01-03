package context

import (
	"context"
	"text/template"

	database "github.com/polyfire/api/db"
	"github.com/polyfire/api/memory"
)

var memoryTemplate = template.Must(
	template.New("memory_context").Parse(`Here are some informations you remember:
{{range .Data}} - {{.}}
{{end}}`),
)

type MemoryContext = TemplateContext

func GetMemory(
	ctx context.Context,
	userID string,
	memoryIDs []string,
	task string,
) (*MemoryContext, error) {
	results := []database.MatchResult{}
	var err error

	if len(memoryIDs) > 0 {
		results, err = memory.Embedder(ctx, userID, memoryIDs, task)
		if err != nil {
			return nil, err
		}
	}

	resultStrings := make([]string, len(results))
	for i, result := range results {
		resultStrings[i] = result.Content
	}

	return GetTemplateContext(resultStrings, *memoryTemplate)
}
