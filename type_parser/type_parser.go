package type_parser

import (
	"errors"
	"fmt"
	"reflect"
)

func nSpaces(indent int) string {
	r := ""
	for i := 0; i < indent; i++ {
		r += " "
	}
	return r
}

func TypeToString(t interface{}, indent int) (string, error) {
	switch v := t.(type) {
	default:
		return "", errors.New(fmt.Sprintf("Unexpected type %T", v))

	case string:
		switch t {
		default:
			return "", errors.New(fmt.Sprintf("Unexpected type %s", t))

		case "string":
			return "string", nil

		case "number":
			return "number", nil

		}

	case []interface{}:
		t := t.([]interface{})

		if len(t) == 0 {
			return "", errors.New("Empty Array")
		}

		sub_t, err := TypeToString(t[0], indent)

		if err != nil {
			return "", err
		}

		s := "[" + sub_t + "]"
		return s, nil

	case map[string]interface{}:
		s := "{\n"
		for key, value := range t.(map[string]interface{}) {
			sub_t, err := TypeToString(value, indent+2)

			if err != nil {
				return "", err
			}

			s += nSpaces(indent+2) + "\"" + key + "\": " + sub_t + ",\n"
		}
		s += nSpaces(indent) + "}"

		return s, nil
	}
}

func CheckAgainstType(t interface{}, v interface{}) bool {
	if t == nil || v == nil {
		return false
	}
	type typePair struct{ t_t, v_t string }
	s := typePair{t_t: reflect.TypeOf(t).String(), v_t: reflect.TypeOf(v).String()}

	switch s {
	default:
		fmt.Printf("AAAAAAAAAAAAAAAAAAAAAAAAAAA\n")
		return false

	case typePair{t_t: "string", v_t: "string"}:
		if t.(string) != "string" {
			fmt.Printf("[[%s]]\n", t.(string))
			return false
		}

	case typePair{t_t: "string", v_t: "float64"}:
		if t.(string) != "number" {
			fmt.Printf("[[%s]]\n", t.(string))
			return false
		}

	case typePair{t_t: "map[string]interface {}", v_t: "map[string]interface {}"}:
		t := t.(map[string]interface{})
		v := v.(map[string]interface{})

		if len(t) != len(v) {
			fmt.Printf("DDDDDDDDDDDDDDDDDDDDDDDDDDDDD\n")
			return false
		}

		for key, t_val := range t {
			if v_val, ok := v[key]; ok {
				if !CheckAgainstType(t_val, v_val) {
					fmt.Printf("BBBBBBBBBBBBBBBBBBBBBBBBBBB\n")
					return false
				}
			} else {
				fmt.Printf("CCCCCCCCCCCCCCCCCCCCCCCCCCC\n")
				return false
			}
		}

	case typePair{t_t: "[]interface {}", v_t: "[]interface {}"}:
		t := t.([]interface{})
		v := v.([]interface{})
		if len(t) != 1 {
			fmt.Printf("FFFFFFFFFFFFFFFFFFFFFFFFFFFFF\n")
			return false
		}

		for _, v_val := range v {
			if !CheckAgainstType(t[0], v_val) {
				fmt.Printf("EEEEEEEEEEEEEEEEEEEEEEEEEEEEE\n")
				return false
			}
		}
	}
	return true
}
