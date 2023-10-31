package completion

import (
	"fmt"
	"strings"

	"github.com/polyfire/api/db"
)

type ParsedSystemPromptElement struct {
	Literal string
	IsVar   bool
}

type SystemPrompt struct {
	elements []ParsedSystemPromptElement
}

func ParseSystemPrompt(system_prompt string) SystemPrompt {
	result := make([]ParsedSystemPromptElement, 0)

	var literal string = ""
	var is_var bool = false

	var last_char rune = 0

	for _, c := range system_prompt {
		if c == '\\' && last_char != '\\' {
			last_char = c
			continue
		} else if c == '{' && last_char == '{' && !is_var {
			result = append(result, ParsedSystemPromptElement{Literal: literal, IsVar: is_var})
			literal = ""
			is_var = true
			last_char = 0
			continue
		} else if c == '}' && last_char == '}' && is_var {
			result = append(result, ParsedSystemPromptElement{Literal: strings.TrimSpace(literal), IsVar: is_var})
			literal = ""
			is_var = false
			last_char = 0
			continue
		} else if last_char == '{' || last_char == '}' {
			last_char = c
			literal += string(last_char) + string(c)
		} else if !(c == '}' || c == '{') || last_char == '\\' {
			last_char = 0
			literal += string(c)
		} else {
			last_char = c
		}
	}

	if last_char == '{' || last_char == '}' {
		literal += string(last_char)
	}

	if is_var {
		literal = "{{" + literal
		is_var = false
	}

	result = append(result, ParsedSystemPromptElement{Literal: literal, IsVar: is_var})

	return SystemPrompt{elements: result}
}

func (sp SystemPrompt) ListVars() []string {
	var result []string = make([]string, 0)

	for _, e := range sp.elements {
		if e.IsVar {
			result = append(result, e.Literal)
		}
	}

	return result
}

func (sp SystemPrompt) Render(vars map[string]string) string {
	var result string = ""

	for _, e := range sp.elements {
		if e.IsVar {
			result += vars[e.Literal]
		} else {
			result += e.Literal
		}
	}

	return result
}

func GetVars(user_id string, varList []string) (map[string]string, []string) {
	var warnings []string = make([]string, 0)
	var result map[string]string = make(map[string]string)

	kv_vars := make([]string, 0)
	for _, v := range varList {
		if strings.HasPrefix(v, "kv.") {
			kv_vars = append(kv_vars, strings.TrimPrefix(v, "kv."))
		}
	}

	kv_map, err := db.GetKVMap(user_id, kv_vars)
	if err != nil {
		fmt.Println(err)
	}

	for _, v := range varList {
		if strings.HasPrefix(v, "kv.") {
			key := strings.TrimPrefix(v, "kv.")
			if kv_map[key] == "" {
				warnings = append(warnings, fmt.Sprintf("Unknown var: \"%s\"", v))
				result[v] = ""
			}
			result[v] = kv_map[key]
		} else {
			warnings = append(warnings, fmt.Sprintf("Unknown var: \"%s\"", v))
			result[v] = ""
		}
	}

	return result, warnings
}

func getSystemPrompt(user_id string, system_prompt_id *string, system_prompt *string) (string, []string, error) {
	var result string = ""

	if system_prompt != nil && len(*system_prompt) > 0 {
		result = *system_prompt
	}

	if system_prompt_id != nil && len(*system_prompt_id) > 0 {
		p, err := db.GetPromptByIdOrSlug(*system_prompt_id)
		if err != nil || p == nil {
			return "", nil, NotFound
		}

		result = p.Prompt
	}

	if len(result) == 0 {
		return result, nil, nil
	}

	systemPrompt := ParseSystemPrompt(result)

	vars, warnings := GetVars(user_id, systemPrompt.ListVars())
	result = systemPrompt.Render(vars)

	if len(warnings) == 0 {
		warnings = nil
	}
	return result, warnings, nil
}
