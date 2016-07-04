package main

import (
	"fmt"
	"regexp"
	"strings"
)

func findElementsInString(p, e string) (string, error) {
	// "(we'd|we would) (like to) (place an)? ?(order) (for|from)? ?(?P<restaurant>.*)"
	re := regexp.MustCompile(p)

	n1 := re.SubexpNames()
	r2 := re.FindAllStringSubmatch("we would like to order chili's", -1)[0]

	matches := map[string]string{}
	for i, n := range r2 {
		matches[n1[i]] = n
	}
	if _, ok := matches[e]; ok {
		return matches[e], nil
	}
	return "", fmt.Errorf("No match found for %v", e)
}

func stringInArray(s string, a []string) bool {
	for _, x := range a {
		if strings.ToLower(s) == x {
			return true
		}
	}
	return false
}

func responseIsAffirmative(response string) bool {
	affirmatives := []string{"yes", "yup", "yep", "sure", "ok"}
	return stringInArray(response, affirmatives)
}

func responseIsNegative(response string) bool {
	negatives := []string{"no", "nope", "cancel", "neup"}
	return stringInArray(response, negatives)
}

func cleanString(s string) string {
	unwanted := regexp.MustCompile("[.,/!$%^*;:{}`=-?\n]")
	doubleSpaces := regexp.MustCompile("  ")
	return strings.ToLower(doubleSpaces.ReplaceAllString(unwanted.ReplaceAllString(s, ""), " "))
}

func matchesPattern(p string, s []string) bool {
	r := regexp.MustCompile(p)
	for _, x := range s {
		if r.Match([]byte(cleanString(x))) {
			return true
		}
	}
	return false
}
