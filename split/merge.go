package split

import (
	"fmt"
	"reflect"
)

func Merge(v1 interface{}, v2 interface{}) interface{} {
	if v1 == nil {
		return v2
	} else if v2 == nil {
		return v1
	}

	type typePair struct{ v1_t, v2_t string }
	s := typePair{v1_t: reflect.TypeOf(v1).String(), v2_t: reflect.TypeOf(v2).String()}

	switch s {
	default:
		panic(fmt.Sprintf("%v\n", s))

	case typePair{v1_t: "string", v2_t: "string"}:
		return v1.(string) + "\n\n" + v2.(string)

	case typePair{v1_t: "[]interface {}", v2_t: "[]interface {}"}:
		return append(v1.([]interface{}), v2.([]interface{})...)

	case typePair{v1_t: "map[string]interface {}", v2_t: "map[string]interface {}"}:
		v1 := v1.(map[string]interface{})
		v2 := v2.(map[string]interface{})
		res := make(map[string]interface{})

		if len(v1) != len(v2) {
			panic("BBBBBBB")
		}

		for key, v1_val := range v1 {
			if v2_val, ok := v2[key]; ok {
				res[key] = Merge(v1_val, v2_val)
			} else {
				panic("CCCCCCCC")
			}
		}
		return res

	case typePair{v1_t: "float64", v2_t: "float64"}:
		return v1.(float64) + v2.(float64)

	case typePair{v1_t: "bool", v2_t: "bool"}:
		return v1.(bool) || v2.(bool)
	}

	return nil
}
