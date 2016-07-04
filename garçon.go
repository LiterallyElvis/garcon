package main

import (
	"fmt"
	"github.com/literallyelvis/slacker"
	"strings"
)

// Garcon is our order taking bot! ヽ(゜∇゜)ノ
type Garcon struct {
	Stage            string
	Interlocutor     slacker.SlackUser
	Order            *Order
	Patrons          map[string]slacker.SlackUser
	MessageTypeFuncs map[string]func(slacker.SlackMessage) (string, error)
	ReactionFuncs    map[string]map[string]func(slacker.SlackMessage) slacker.SlackMessage
}

// Reset wipes the state of Garcon
func (g *Garcon) Reset() {
	g.Stage = "uninitiated"
	g.Order = NewOrder()
}

// NewGarcon constructs a new instance of Garcon and establishes all the behavior functions
// Note that while it is possible for us to make this code broader and more reusable, I don't
// have an intense interest in doing that right now, and I think such a structure would render
// the code either hilariously unreadable, complicated, and most likely both
func NewGarcon() *Garcon {
	g := &Garcon{}
	g.Stage = "uninitiated"

	// possible returns: affirmative, negative, additive, cancelling, irrelevant
	g.MessageTypeFuncs = map[string]func(slacker.SlackMessage) (string, error){
		"uninitiated": func(m slacker.SlackMessage) (string, error) {
			if strings.ToLower(m.Text) == "oh, garçon?" && !g.Order.Begun {
				// ai.Order.Interlocutor = m.User
				return "affirmative", nil
			}
			return "irrelevant", nil
		},
		"prompted": func(m slacker.SlackMessage) (string, error) {
			return "", nil
		},
		"ordering": func(m slacker.SlackMessage) (string, error) {
			return "", nil
		},
		"confirmation": func(m slacker.SlackMessage) (string, error) {
			return "", nil
		},
	}

	g.ReactionFuncs = map[string]map[string]func(slacker.SlackMessage) slacker.SlackMessage{
		"uninitiated": map[string]func(m slacker.SlackMessage) slacker.SlackMessage{
			"affirmative": func(m slacker.SlackMessage) slacker.SlackMessage {
				t := fmt.Sprintf("Hi, @%v! Would you like to place an order?", g.Patrons[m.User].Username)
				g.Stage = "prompted"
				return slacker.SlackMessage{Channel: m.Channel, Text: t}
			},
		},
		"prompted": map[string]func(m slacker.SlackMessage) slacker.SlackMessage{
			"cancelling": func(m slacker.SlackMessage) slacker.SlackMessage {
				t := "Okay then, I'll disappear for now"
				g.Reset()
				return slacker.SlackMessage{Channel: m.Channel, Text: t}
			},
		},
		"ordering": map[string]func(m slacker.SlackMessage) slacker.SlackMessage{
			"cancelling": func(m slacker.SlackMessage) slacker.SlackMessage {
				t := "Okay then, I'll disappear for now"
				g.Reset()
				return slacker.SlackMessage{Channel: m.Channel, Text: t}
			},
		},
		"confirmation": map[string]func(m slacker.SlackMessage) slacker.SlackMessage{
			"cancelling": func(m slacker.SlackMessage) slacker.SlackMessage {
				t := "Okay then, I'll disappear for now"
				g.Reset()
				return slacker.SlackMessage{Channel: m.Channel, Text: t}
			},
		},
	}

	return g
}
