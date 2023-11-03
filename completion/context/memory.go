package context

import (
	"context"
	"text/template"

	"github.com/polyfire/api/memory"
)

var MEMORY_TEMPLATE = template.Must(template.New("memory_context").Parse(`Here are some informations you remember:
{{range .Data}} - {{.}}
{{end}}`))

type MemoryContext = TemplateContext

func GetMemory(ctx context.Context, userId string, memoryIds []string, task string) (*MemoryContext, error) {
	results, err := memory.Embedder(ctx, userId, memoryIds, task)
	if err != nil {
		return nil, err
	}

	resultStrings := make([]string, len(results))
	for i, result := range results {
		resultStrings[i] = result.Content
	}

	return GetTemplateContext(resultStrings, *MEMORY_TEMPLATE)
}
