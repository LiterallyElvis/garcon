package main

import (
	"regexp"
	"strings"
)

func findElementsInString() {
	re := regexp.MustCompile("(we'd|we would) (like to) (place an)? ?(order) (for|from)? ?(?P<restaurant>.*)")

	n1 := re.SubexpNames()
	r2 := re.FindAllStringSubmatch("we would like to order chili's", -1)[0]

	md := map[string]string{}
	for i, n := range r2 {
		md[n1[i]] = n
	}
	// fmt.Printf("The restauraunt is %s\n", md["restaurant"])
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
