package type_parser

import (
	"testing"
)

func TestTypeToString1(t *testing.T) {
	tata := make(map[string]interface{})
	tata["test"] = "string"
	got, err := TypeToString(tata, 0)
	want := "{\n  \"test\": string,\n}"

	if err != nil {
		t.Errorf("error (%v)", err)
	} else if got != want {
		t.Errorf("got (%v) wanted (%v)", got, want)
	}
}

func TestTypeToString2(t *testing.T) {
	tata := make(map[string]interface{})
	tutu := make(map[string]interface{})
	tata["test"] = "string"
	tutu["e"] = "boolean"
	tata["i"] = tutu
	got, err := TypeToString(tata, 0)
	want := "{\n  \"test\": string,\n  \"i\": {\n    \"e\": boolean,\n  },\n}"
	want2 := "{\n  \"i\": {\n    \"e\": boolean,\n  },\n  \"test\": string,\n}"

	if err != nil {
		t.Errorf("error (%v)", err)
	} else if got != want && got != want2 {
		t.Errorf("got (%v) wanted (%v)", got, want)
	}
}

func TestCheckAgainstType1(t *testing.T) {
	tata := make(map[string]interface{})
	tutu := make(map[string]interface{})
	tata["test"] = "string"
	tutu["e"] = "boolean"
	tata["i"] = tutu

	t1 := make(map[string]interface{})
	t2 := make(map[string]interface{})
	t1["test"] = "abc"
	t2["e"] = true
	t1["i"] = t2

	got := CheckAgainstType(tata, t1)
	want := true

	if got != want {
		t.Errorf("got (%v) wanted (%v)", got, want)
	}
}

func TestCheckAgainstType2(t *testing.T) {
	tata := make(map[string]interface{})
	tutu := make(map[string]interface{})
	tata["test"] = "string"
	tutu["e"] = "number"
	tata["i"] = tutu

	t1 := make(map[string]interface{})
	t2 := make(map[string]interface{})
	t1["test"] = "abc"
	t2["e"] = true
	t1["i"] = t2

	got := CheckAgainstType(tata, t1)
	want := false

	if got != want {
		t.Errorf("got (%v) wanted (%v)", got, want)
	}
}
