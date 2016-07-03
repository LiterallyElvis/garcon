package main

import (
	"testing"
)

// Testing creation of the header to index map
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
