package context

import (
	"errors"
	"fmt"
	"strings"

	"github.com/polyfire/api/db"
	"github.com/polyfire/api/tokens"
)

type ParsedSystemPromptElement struct {
	Literal string
	IsVar   bool
}

type SystemPrompt struct {
	elements []ParsedSystemPromptElement
}

func ParseSystemPrompt(systemPrompt string) SystemPrompt {
	result := make([]ParsedSystemPromptElement, 0)

	literal := ""
	isVar := false

	var lastChar rune

	for _, c := range systemPrompt {
		if c == '\\' && lastChar != '\\' {
			lastChar = c
			continue
		} else if c == '{' && lastChar == '{' && !isVar {
			result = append(result, ParsedSystemPromptElement{Literal: literal, IsVar: isVar})
			literal = ""
			isVar = true
			lastChar = 0
			continue
		} else if c == '}' && lastChar == '}' && isVar {
			result = append(result, ParsedSystemPromptElement{Literal: strings.TrimSpace(literal), IsVar: isVar})
			literal = ""
			isVar = false
			lastChar = 0
			continue
		} else if lastChar == '{' || lastChar == '}' {
			lastChar = c
			literal += string(lastChar) + string(c)
		} else if !(c == '}' || c == '{') || lastChar == '\\' {
			lastChar = 0
			literal += string(c)
		} else {
			lastChar = c
		}
	}

	if lastChar == '{' || lastChar == '}' {
		literal += string(lastChar)
	}

	if isVar {
		literal = "{{" + literal
		isVar = false
	}

	result = append(result, ParsedSystemPromptElement{Literal: literal, IsVar: isVar})

	return SystemPrompt{elements: result}
}

func (sp SystemPrompt) ListVars() []string {
	var result = make([]string, 0)

	for _, e := range sp.elements {
		if e.IsVar {
			result = append(result, e.Literal)
		}
	}

	return result
}

func (sp SystemPrompt) Render(vars map[string]string) string {
	var result= ""

	for _, e := range sp.elements {
		if e.IsVar {
			result += vars[e.Literal]
		} else {
			result += e.Literal
		}
	}

	return result
}

func GetVars(userID string, varList []string) (map[string]string, []string) {
	var warnings = make([]string, 0)
	var result = make(map[string]string)

	kvVars := make([]string, 0)
	for _, v := range varList {
		if strings.HasPrefix(v, "kv.") {
			kvVars = append(kvVars, strings.TrimPrefix(v, "kv."))
		}
	}

	kvMap, err := db.GetKVMap(userID, kvVars)
	if err != nil {
		fmt.Println(err)
	}

	for _, v := range varList {
		if strings.HasPrefix(v, "kv.") {
			key := strings.TrimPrefix(v, "kv.")
			if kvMap[key] == "" {
				warnings = append(warnings, fmt.Sprintf("Unknown var: \"%s\"", v))
				result[v] = ""
			}
			result[v] = kvMap[key]
		} else {
			warnings = append(warnings, fmt.Sprintf("Unknown var: \"%s\"", v))
			result[v] = ""
		}
	}

	return result, warnings
}

type SystemPromptContext struct {
	SystemPrompt string
}

func GetSystemPrompt(
	userID string,
	systemPromptID *string,
	systemPrompt *string,
	chatID *string,
) (*SystemPromptContext, []string, error) {
	var result = ""

	if systemPrompt != nil && len(*systemPrompt) > 0 {
		result = *systemPrompt
	}

	if chatID != nil && len(*chatID) > 0 {
		c, err := db.GetChatByID(*chatID)
		if err != nil {
			return nil, nil, errors.New("Chat not found")
		}

		if c.SystemPromptID != nil && len(*c.SystemPromptID) > 0 {
			systemPromptID = c.SystemPromptID
		}
	}

	if systemPromptID != nil && len(*systemPromptID) > 0 {
		p, err := db.GetPromptByIDOrSlug(*systemPromptID)
		if err != nil || p == nil {
			return nil, nil, errors.New("Prompt not found")
		}

		result = p.Prompt
	}

	if len(result) == 0 {
		return nil, nil, errors.New("No prompt provided")
	}

	systemPromptCtx := ParseSystemPrompt(result)

	varList := systemPromptCtx.ListVars()

	var warnings []string

	if len(varList) != 0 {
		var vars map[string]string
		vars, warnings = GetVars(userID, varList)
		result = systemPromptCtx.Render(vars)

		if len(warnings) == 0 {
			warnings = nil
		}
	}
	return &SystemPromptContext{SystemPrompt: result + "\n"}, warnings, nil
}

func (spc *SystemPromptContext) GetOrderIndex() int {
	return 1
}

func (spc *SystemPromptContext) GetPriority() Priority {
	return CRITICAL
}

func (spc *SystemPromptContext) GetMinimumContextSize() int {
	return tokens.CountTokens(spc.SystemPrompt)
}

func (spc *SystemPromptContext) GetRecommendedContextSize() int {
	return tokens.CountTokens(spc.SystemPrompt)
}

func (spc *SystemPromptContext) GetContentFittingIn(tokenCount int) string {
	if tokens.CountTokens(spc.SystemPrompt) > tokenCount {
		return ""
	}
	return spc.SystemPrompt
}
