package main

import (
	"fmt"
	"github.com/literallyelvis/slacker"
	"log"
	"strings"
)

// Garcon is our order taking bot! ヽ(゜∇゜)ノ
type Garcon struct {
	debug            bool
	Stage            string
	InterlocutorID   string
	Order            *Order
	Patrons          map[string]slacker.SlackUser
	MessageTypeFuncs map[string]func(slacker.SlackMessage) (string, error)
	ReactionFuncs    map[string]map[string]func(slacker.SlackMessage) []slacker.SlackMessage
}

// Reset wipes the state of Garcon
func (g *Garcon) Reset() {
	g.Stage = "uninitiated"
	g.Order = NewOrder()
}

// CancellationCommandIssued returns whether or not the most recent command is a show-stopping
// cancellation command directed at Garcon
func (g Garcon) CancellationCommandIssued(m string) (abortCommandIssued bool) {
	log.Printf("CancellationCommandIssued called against %v\n", m)
	if stringFitsPattern("(<@\\w+>, abort)", m) {
		match, _ := findElementsInString("(<@(?P<user>\\w+)>, abort)", "user", "<@U1N3QR9F1>, abort")
		if _, ok := g.Patrons[match]; ok {
			log.Printf("g.Patrons[match]: %v\n", g.Patrons[match].Username)
			if strings.ToLower(g.Patrons[match].Username) == "garcon" {
				log.Println("strings.ToLower(g.Patrons[match].Username) == \"garcon\"")
				abortCommandIssued = true
			}
		}
	}

	if g.debug {
		log.Printf("Checked if message received was cancellation command. Returning %v\n", abortCommandIssued)
	}
	return abortCommandIssued
}

// NewGarcon constructs a new instance of Garcon and establishes all the behavior functions
// Note that while it is possible for us to make this code broader and more reusable, I don't
// have an intense interest in doing that right now, and I think such a structure would render
// the code either hilariously unreadable, complicated, and most likely both
func NewGarcon() *Garcon {
	g := &Garcon{}
	g.Stage = "uninitiated"

	// possible returns: affirmative, negative, additive, cancelling, irrelevant, indeterminable
	g.MessageTypeFuncs = map[string]func(slacker.SlackMessage) (string, error){
		"uninitiated": func(m slacker.SlackMessage) (string, error) {
			if g.CancellationCommandIssued(m.Text) {
				return "cancelling", nil
			}
			if strings.ToLower(m.Text) == "oh, garçon?" {
				return "affirmative", nil
			}
			return "irrelevant", nil
		},
		"prompted": func(m slacker.SlackMessage) (string, error) {
			if g.CancellationCommandIssued(m.Text) {
				return "cancelling", nil
			}
			negative := responseIsNegative(m.Text)
			if negative || m.User != g.InterlocutorID {
				return "negative", nil
			}
			restaurant, err := findElementsInString("(we'd|we would) (like to) (place an)? ?(order) (for|from)? ?(?P<restaurant>.*)", "restauraunt", m.Text)
			if err != nil {
				return "indeterminable", err
			}
			if restaurant != "" {
				return "affirmative", nil
			}
			return "indeterminable", nil
		},
		"ordering": func(m slacker.SlackMessage) (string, error) {
			if g.CancellationCommandIssued(m.Text) {
				return "cancelling", nil
			}
			return "indeterminable", nil
		},
		"confirmation": func(m slacker.SlackMessage) (string, error) {
			if g.CancellationCommandIssued(m.Text) {
				return "cancelling", nil
			}
			return "indeterminable", nil
		},
	}

	genericCancelReponse := func(m slacker.SlackMessage) []slacker.SlackMessage {
		g.Reset()
		t := "Very well then, I'll disappear for now!"
		return []slacker.SlackMessage{slacker.SlackMessage{Channel: m.Channel, Text: t}}
	}

	genericHelpResponse := func(m slacker.SlackMessage, ec []string) []slacker.SlackMessage {
		t := fmt.Sprintf("I'm sorry, @%v, I couldn't understand what you said. Here are some things I might understand:\n%v\n", g.Patrons[m.User], strings.Join(ec, "\n"))
		return []slacker.SlackMessage{slacker.SlackMessage{Channel: m.Channel, Text: t}}
	}

	g.ReactionFuncs = map[string]map[string]func(slacker.SlackMessage) []slacker.SlackMessage{
		"uninitiated": map[string]func(m slacker.SlackMessage) []slacker.SlackMessage{
			"affirmative": func(m slacker.SlackMessage) []slacker.SlackMessage {
				t := fmt.Sprintf("Hi, @%v! Would you like to place an order?", g.Patrons[m.User].Username)
				g.InterlocutorID = m.User
				g.Stage = "prompted"

				return []slacker.SlackMessage{
					slacker.SlackMessage{Channel: "C1N3MEUMN", Text: t},
				}
			},
		},
		"prompted": map[string]func(m slacker.SlackMessage) []slacker.SlackMessage{
			"affirmative": func(m slacker.SlackMessage) []slacker.SlackMessage {
				exampleResponses := []string{
					"We'd like to place an order from the Chili's at 45th & Lamar",
					"We would like to order from the Chili's at 45th & Lamar",
				}
				restaurant, err := findElementsInString("(we'd|we would) (like to) (place an)? ?(order) (for|from)? ?(?P<restaurant>.*)", "restauraunt", m.Text)
				if err != nil || len(restaurant) == 0 {
					return genericHelpResponse(m, exampleResponses)
				}
				t := fmt.Sprintf("Okay, what would you like from %v?", restaurant)
				return []slacker.SlackMessage{slacker.SlackMessage{Channel: m.Channel, Text: t}}
			},
			"cancelling": genericCancelReponse,
		},
		"ordering": map[string]func(m slacker.SlackMessage) []slacker.SlackMessage{
			"cancelling": genericCancelReponse,
		},
		"confirmation": map[string]func(m slacker.SlackMessage) []slacker.SlackMessage{
			"cancelling": genericCancelReponse,
		},
	}

	return g
}
