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

func TestMatchesPattern(t *testing.T) {
	m := stringFitsPattern("(<@\\w+>, abort)", "<@LOLWTFBBQ>, abort")
	if !m {
		t.Errorf("stringFitsPattern failed to detect valid string in valid pattern")
	}

}

// Testing creation of the header to index map
func TestStringElementFinding(t *testing.T) {
	match, err := findElementsInString("(we'd|we would) (like to) (place an)? ?(order) (for|from)? ?(?P<restaurant>.*)", "restaurant", "We'd like to place an order from the Chili's at 45th & Lamar")
	if err != nil {
		t.Errorf("Unexpected error encountered when trying to find elements in string:\n%v\n", err)
	}
	if match != "the Chili's at 45th & Lamar" {
		t.Errorf("expected matched string to equal \"the Chili's on 45th & Lamar\", instead equaled \"%v\"", match)
	}

	match, err = findElementsInString("(<@(?P<user>\\w+)>, abort)", "user", "<@U1N3QR9F1>, abort")
	if err != nil {
		t.Errorf("Unexpected error encountered when trying to find elements in string:\n%v\n", err)
	}
	if match != "U1N3QR9F1" {
		t.Errorf("expected matched string to equal \"U1N3QR9F1\", instead equaled \"%v\"", match)
	}
}
