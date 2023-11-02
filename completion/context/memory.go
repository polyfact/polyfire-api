package context

import (
	"bytes"
	"context"
	"text/template"

	"github.com/polyfire/api/db"
	"github.com/polyfire/api/memory"

	"github.com/polyfire/api/tokens"
)

var MEMORY_TEMPLATE = template.Must(template.New("memory_context").Parse(`Memory:
{{range .Memory}} - {{.}}
{{end}}`))

type MemoryTemplateData struct {
	Memory []string
}

type MemoryTemplateGrowth struct {
	// The growth of the number of tokens in the template in relation to the number of memories
	// Of the form: A * x + B
	A int
	B int
}

func InitContextStructureTemplate() MemoryTemplateGrowth {
	var result MemoryTemplateGrowth

	data1 := MemoryTemplateData{Memory: []string{}}
	data2 := MemoryTemplateData{Memory: []string{""}}

	var result1 bytes.Buffer
	var result2 bytes.Buffer

	if err := MEMORY_TEMPLATE.Execute(&result1, data1); err != nil {
		panic(err)
	}
	if err := MEMORY_TEMPLATE.Execute(&result2, data2); err != nil {
		panic(err)
	}

	result.B = tokens.CountTokens("gpt-3.5-turbo", result1.String())
	result.A = tokens.CountTokens("gpt-3.5-turbo", result2.String()) - result.B

	return result
}

var MEMORY_TEMPLATE_TOKENS_GROWTH = InitContextStructureTemplate()

type MemoryContext struct {
	MemoryIds    []string
	MatchResults []db.MatchResult
}

func GetMemory(ctx context.Context, userId string, memoryIds []string, task string) (*MemoryContext, error) {
	results, err := memory.Embedder(ctx, userId, memoryIds, task)
	if err != nil {
		return nil, err
	}

	memoryContext := MemoryContext{
		MemoryIds:    memoryIds,
		MatchResults: results,
	}

	return &memoryContext, nil
}

func (m *MemoryContext) GetMinimumContextSize() int {
	if len(m.MatchResults) == 0 {
		return 0
	}

	return MEMORY_TEMPLATE_TOKENS_GROWTH.A + tokens.CountTokens(
		"gpt-3.5-turbo",
		m.MatchResults[0].Content,
	) + MEMORY_TEMPLATE_TOKENS_GROWTH.B
}

func (m *MemoryContext) GetRecommendedContextSize() int {
	if len(m.MatchResults) == 0 {
		return 0
	}

	totalTokens := MEMORY_TEMPLATE_TOKENS_GROWTH.B

	for _, item := range m.MatchResults {
		totalTokens += tokens.CountTokens("gpt-3.5-turbo", item.Content) + MEMORY_TEMPLATE_TOKENS_GROWTH.A
	}

	return totalTokens
}

func (m *MemoryContext) GetPriority() Priority {
	return HELPFUL
}

func fillContext(embeddings []db.MatchResult, token_count int) (string, error) {
	memories := []string{}
	currentTokens := MEMORY_TEMPLATE_TOKENS_GROWTH.B

	for _, item := range embeddings {
		textTokens := tokens.CountTokens("gpt3.5-turbo", item.Content)

		if currentTokens+textTokens+MEMORY_TEMPLATE_TOKENS_GROWTH.A > token_count {
			break
		}

		memories = append(memories, item.Content)
		currentTokens += textTokens + MEMORY_TEMPLATE_TOKENS_GROWTH.A
	}

	if len(memories) == 0 {
		return "", nil
	}

	data := MemoryTemplateData{
		Memory: memories,
	}

	var result bytes.Buffer

	if err := MEMORY_TEMPLATE.Execute(&result, data); err != nil {
		return "", err
	}

	context := result.String()

	return context, nil
}

func (m *MemoryContext) GetContentFittingIn(token_count int) string {
	context, _ := fillContext(m.MatchResults, token_count)
	return context
}
