package main

import (
	"fmt"
	"regexp"
	"strings"
)

func stringFitsPattern(p, s string) bool {
	re := regexp.MustCompile(p)
	return re.Match([]byte(s))
}

func findElementsInString(p string, e []string, m string) (map[string]string, error) {
	re := regexp.MustCompile(fmt.Sprintf("(?i)%v", p))

	n1 := re.SubexpNames()
	r2 := re.FindAllStringSubmatch(m, -1)
	if len(r2) == 0 {
		return nil, fmt.Errorf("No string submatches found for %v", e)
	}
	r3 := r2[0]

	matches := map[string]string{}
	for i, n := range r3 {
		matches[n1[i]] = n
	}
	returnMatches := map[string]string{}
	for _, el := range e {
		if _, ok := matches[el]; ok {
			returnMatches[el] = matches[el]
		}
	}
	if len(returnMatches) > 0 {
		return returnMatches, nil
	}

	return nil, fmt.Errorf("No match found for %v", e)
}

func stringInArray(s string, a []string) bool {
	for _, x := range a {
		if cleanString(s) == x {
			return true
		}
	}
	return false
}

func responseIsAffirmative(response string) bool {
	affirmatives := []string{"yes", "yup", "yep", "sure", "ok", "si", "oui"}
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
