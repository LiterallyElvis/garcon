package main

import (
	"testing"
)

func TestStringCleaning(t *testing.T) {
	badStrings := map[string]string{
		"Whoa!":            "whoa",
		"What?!":           "what",
		"things; stuff":    "things stuff",
		"things?\n stuff!": "things stuff",
		"{`more`  things}": "more things",
	}

	for d, expected := range badStrings {
		actual := cleanString(d)
		if expected != actual {
			t.Errorf("Dirty string not cleaned!\n\tExpected: %v\n\tReceived: %v\n\n", expected, actual)
		}
	}
}

func TestStringInSlice(t *testing.T) {
	dummySlice := []string{
		"things",
		"and",
		"stuff",
	}
	correct := sliceContainsString("things", dummySlice)
	if !correct {
		t.Error("String was determined not to be present in array, when it should have been.")
	}
}

func TestMatchesPattern(t *testing.T) {
	m := stringFitsPattern("(<@\\w+>, abort)", "<@LOLWTFBBQ>, abort")
	if !m {
		t.Errorf("stringFitsPattern failed to detect valid string in valid pattern")
	}
}
