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

func TestMatchesPattern(t *testing.T) {
	m := stringFitsPattern("(<@\\w+>, abort)", "<@LOLWTFBBQ>, abort")
	if !m {
		t.Errorf("stringFitsPattern failed to detect valid string in valid pattern")
	}
}

func TestRestaurantFinding(t *testing.T) {
	restaurantMatches, err := findElementsInString(orderInitiationPattern, []string{"restaurant"}, "We'd like to place an order from the Chili's at 45th & Lamar")
	expectedRestaurant := "the Chili's at 45th & Lamar"
	actualRestaurant := restaurantMatches["restaurant"]

	if err != nil {
		t.Errorf("Unexpected error encountered when trying to find elements in string:\n%v\n", err)
	}
	if actualRestaurant != expectedRestaurant {
		t.Errorf("expected matched string to equal %v, instead equaled \"%v\"", expectedRestaurant, actualRestaurant)
	}
}

func TestOrderPlacementFinding(t *testing.T) {
	userOrderMatches, err := findElementsInString(orderPlacingPattern, []string{"user", "item"}, "<@U1N3QR9F1>, I would like a thing")
	expectedUser := "U1N3QR9F1"
	actualUser := userOrderMatches["user"]
	expectedItem := "a thing"
	actualItem := userOrderMatches["item"]
	if err != nil {
		t.Errorf("Unexpected error encountered when trying to find elements in string:\n%v\n", err)
	}
	if expectedUser != actualUser {
		t.Errorf("expected matched string to equal %v, instead equaled \"%v\"", expectedUser, actualUser)
	}
	if expectedItem != actualItem {
		t.Errorf("expected matched string to equal %v, instead equaled \"%v\"", expectedItem, actualItem)
	}

	userOrderMatches, err = findElementsInString(orderPlacingPattern, []string{"user", "item"}, "<@U1N3QR9F1> I'll have the tuna melt")
	expectedUser = "U1N3QR9F1"
	actualUser = userOrderMatches["user"]
	expectedItem = "the tuna melt"
	actualItem = userOrderMatches["item"]
	if err != nil {
		t.Errorf("Unexpected error encountered when trying to find elements in string:\n%v\n", err)
	}
	if expectedUser != actualUser {
		t.Errorf("expected matched string to equal %v, instead equaled \"%v\"", expectedUser, actualUser)
	}
	if expectedItem != actualItem {
		t.Errorf("expected matched string to equal %v, instead equaled \"%v\"", expectedItem, actualItem)
	}
}
