package completion

import (
	"fmt"

	"github.com/polyfire/api/db"
)

type ParsedSystemPromptElement struct {
	Literal string
	IsVar   bool
}

type SystemPrompt = []ParsedSystemPromptElement

func ParseSystemPrompt(system_prompt string) SystemPrompt {
	var result SystemPrompt = make([]ParsedSystemPromptElement, 0)

	var literal string = ""
	var is_var bool = false

	var last_char rune = 0

	for _, c := range system_prompt {
		if c == '\\' && last_char != '\\' {
			last_char = c
			continue
		} else if c == '{' && last_char == '{' {
			result = append(result, ParsedSystemPromptElement{Literal: literal, IsVar: is_var})
			literal = ""
			is_var = true
			last_char = 0
			continue
		} else if c == '}' && last_char == '}' {
			result = append(result, ParsedSystemPromptElement{Literal: literal, IsVar: is_var})
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

		fmt.Println(string(c))
		fmt.Println(literal)

	}

	if is_var {
		literal = "{{" + literal
		is_var = false
	}
	result = append(result, ParsedSystemPromptElement{Literal: literal, IsVar: is_var})

	return result
}

func getSystemPrompt(user_id string, system_prompt_id *string, system_prompt *string) (string, error) {
	var result string = ""

	if system_prompt != nil && len(*system_prompt) > 0 {
		result = *system_prompt
	}

	if system_prompt_id != nil && len(*system_prompt_id) > 0 {
		p, err := db.GetPromptByIdOrSlug(*system_prompt_id)
		if err != nil || p == nil {
			return "", NotFound
		}

		result = p.Prompt
	}

	return result, nil
}
