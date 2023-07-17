package split

import (
	"reflect"
	"testing"
)

func TestMerge1(t *testing.T) {
	v1 := "HELLO"
	v2 := "IT'S ME"
	want := "HELLO\n\nIT'S ME"

	got := Merge(v1, v2)

	if got != want {
		t.Errorf("got (%v) wanted (%v)", got, want)
	}
}

func TestMerge2(t *testing.T) {
	v1 := make([]interface{}, 3)
	v2 := make([]interface{}, 2)
	want := make([]interface{}, 5)
	want[0] = 1.0
	v1[0] = 1.0
	want[1] = 2.0
	v1[1] = 2.0
	want[2] = 3.0
	v1[2] = 3.0
	want[3] = 4.0
	v2[0] = 4.0
	want[4] = 5.0
	v2[1] = 5.0

	got := Merge(v1, v2)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got (%v) wanted (%v)", got, want)
	}
}

func TestMerge3(t *testing.T) {
	v1 := make(map[string]interface{})
	v2 := make(map[string]interface{})
	want := make(map[string]interface{})

	v1["a"] = "HELLO"
	v2["a"] = "IT'S ME"
	want["a"] = "HELLO\n\nIT'S ME"

	got := Merge(v1, v2)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got (%v) wanted (%v)", got, want)
	}
}

func TestMerge4(t *testing.T) {
	v1 := 2.0
	v2 := 3.0
	want := 5.0

	got := Merge(v1, v2)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got (%v) wanted (%v)", got, want)
	}
}

func TestMerge5(t *testing.T) {
	v1 := true
	v2 := false
	want := true

	got := Merge(v1, v2)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got (%v) wanted (%v)", got, want)
	}
}
