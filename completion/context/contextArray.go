package context

import (
	"bytes"
	"text/template"

	"github.com/polyfire/api/tokens"
)

type TemplateData struct {
	Data []string
}

type TemplateGrowth struct {
	// The growth of the number of tokens in the template in relation to the number of memories
	// Of the form: A * x + B
	A int
	B int
}

func InitContextStructureTemplate(templ template.Template) TemplateGrowth {
	var result TemplateGrowth

	data1 := TemplateData{Data: []string{}}
	data2 := TemplateData{Data: []string{""}}

	var result1 bytes.Buffer
	var result2 bytes.Buffer

	if err := templ.Execute(&result1, data1); err != nil {
		panic(err)
	}
	if err := templ.Execute(&result2, data2); err != nil {
		panic(err)
	}

	result.B = tokens.CountTokens(result1.String())
	result.A = tokens.CountTokens(result2.String()) - result.B

	return result
}

type TemplateContext struct {
	Data          []string
	Template      template.Template
	ContextGrowth TemplateGrowth
}

func GetTemplateContext(data []string, templ template.Template) (*TemplateContext, error) {
	memoryContext := TemplateContext{
		Data:          data,
		Template:      templ,
		ContextGrowth: InitContextStructureTemplate(templ),
	}

	return &memoryContext, nil
}

func (m *TemplateContext) GetMinimumContextSize() int {
	if len(m.Data) == 0 {
		return 0
	}

	return m.ContextGrowth.A + tokens.CountTokens(
		m.Data[0],
	) + m.ContextGrowth.B
}

func (m *TemplateContext) GetRecommendedContextSize() int {
	if len(m.Data) == 0 {
		return 0
	}

	totalTokens := m.ContextGrowth.B

	for _, item := range m.Data {
		totalTokens += tokens.CountTokens(item) + m.ContextGrowth.A
	}

	return totalTokens
}

func (m *TemplateContext) GetPriority() Priority {
	return HELPFUL
}

func (m *TemplateContext) GetOrderIndex() int {
	return 2
}

func (m *TemplateContext) fillContext(data []string, tokenCount int) (string, error) {
	memories := []string{}
	currentTokens := m.ContextGrowth.B

	for _, item := range data {
		textTokens := tokens.CountTokens(item)

		if currentTokens+textTokens+m.ContextGrowth.A > tokenCount {
			break
		}

		memories = append(memories, item)
		currentTokens += textTokens + m.ContextGrowth.A
	}

	if len(memories) == 0 {
		return "", nil
	}

	templData := TemplateData{
		Data: memories,
	}

	var result bytes.Buffer

	if err := m.Template.Execute(&result, templData); err != nil {
		return "", err
	}

	context := result.String()

	return context, nil
}

func (m *TemplateContext) GetContentFittingIn(tokenCount int) string {
	context, _ := m.fillContext(m.Data, tokenCount)
	return context
}
