package split

import (
	"reflect"
	"testing"
)

func TestSplitSimple(t *testing.T) {
	got1, got2 := Split("hellomynameislancelot")
	want1 := "hellomynam"
	want2 := "eislancelot"

	if got1 != want1 || got2 != want2 {
		t.Errorf("got (%v, %v) wanted (%v, %v)", got1, got2, want1, want2)
	}
}

func TestSplitSimple2(t *testing.T) {
	got1, got2 := Split("hello,\nmynameislancelot")
	want1 := "hello,"
	want2 := "\nmynameislancelot"

	if got1 != want1 || got2 != want2 {
		t.Errorf("got (%v, %v) wanted (%v, %v)", got1, got2, want1, want2)
	}
}

func TestSplitSimple3(t *testing.T) {
	got1, got2 := Split("hello,\n\nmynamei\nslancelot")
	want1 := "hello,"
	want2 := "\n\nmynamei\nslancelot"

	if got1 != want1 || got2 != want2 {
		t.Errorf("got (%v, %v) wanted (%v, %v)", got1, got2, want1, want2)
	}
}

func TestSplitSimple4(t *testing.T) {
	got1, got2 := Split("hello,\n\n\tmynam\n\n\teislancelot")
	want1 := "hello,"
	want2 := "\n\n\tmynam\n\n\teislancelot"

	if got1 != want1 || got2 != want2 {
		t.Errorf("got (%v, %v) wanted (%v, %v)", got1, got2, want1, want2)
	}
}

func TestBinarySplit(t *testing.T) {
	res := BinarySplit("According to all known laws of aviation, there is no way that a bee should be able to fly.\nIts wings are too small to get its fat little body off the ground.\nThe bee, of course, flies anyway. Because bees don’t care what humans think is impossible.”\n\nSEQ. 75 - “INTRO TO BARRY”\nINT. BENSON HOUSE - DAY ANGLE ON: Sneakers on the ground.\nCamera PANS UP to reveal BARRY BENSON’S BEDROOM ANGLE ON:", 10)
	want := []string{
		"According to all known laws of aviation, there",
		" is no way that a bee",
		" should be able to fly.",
		"\nIts wings are too small to get its",
		" fat little body off the ground.",
		"\nThe bee, of course, flies",
		" anyway. Because bees",
		" don’t care what humans think is impossible.”",
		"\n\nSEQ. 75 - “INTRO",
		" TO BARRY”",
		"\nINT. BENSON",
		" HOUSE - DAY ANGLE",
		" ON: Sneakers on the ground.",
		"\nCamera PANS UP to reveal BARRY",
		" BENSON’S BEDROOM ANGLE ON:",
	}

	if !reflect.DeepEqual(res, want) {
		t.Errorf("got (%v) wanted (%v)", res, want)
	}
}
