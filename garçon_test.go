package main

import (
	// "github.com/literallyelvis/solid"
	// "github.com/nlopes/slack"
	"testing"
)

func TestGarconConstructsWithoutError(t *testing.T) {
	g := NewGarcon()
	g.Stage = "test"
}
