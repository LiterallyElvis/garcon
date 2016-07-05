package main

import (
	"fmt"
	"github.com/nlopes/slack"
	"log"
	"strings"
)

const (
	orderInitiationPattern = "(we'd|we would) (like to) (place an)? ?(order) (for|from)? ?(?P<restaurant>.*)"
)

// Garcon is our order taking bot! ヽ(゜∇゜)ノ
type Garcon struct {
	debug            bool
	Stage            string
	InterlocutorID   string
	Order            *Order
	Patrons          map[string]slack.User
	MessageTypeFuncs map[string]func(slack.Msg) (string, error)
	ReactionFuncs    map[string]map[string]func(slack.Msg) []slack.OutgoingMessage
}

// Reset wipes the state of Garcon
func (g *Garcon) Reset() {
	g.Stage = "uninitiated"
	g.Order = NewOrder()
}

// CancellationCommandIssued returns whether or not the most recent command is a show-stopping
// cancellation command directed at Garcon
func (g Garcon) CancellationCommandIssued(m string) (abortCommandIssued bool) {
	if stringFitsPattern("(<@\\w+>, abort)", m) {
		match, _ := findElementsInString("(<@(?P<user>\\w+)>, abort)", []string{"user"}, m)
		user := match["user"]
		if _, ok := g.Patrons[user]; ok {
			if strings.ToLower(g.Patrons[user].Name) == "garcon" {
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

// ItemAddedToOrder TODO: Document
func (g Garcon) ItemAddedToOrder(m string) (itemAdded bool) {
	if stringFitsPattern("(<@\\w+>, I would like .*)", m) {
		matches, _ := findElementsInString("(<@(?P<user>\\w+)>, I would like (?P<item>.*))", []string{"user", "item"}, m)
		if _, ok := g.Patrons[matches["user"]]; ok {
			if strings.ToLower(g.Patrons[matches["user"]].Name) == "garcon" && len(matches["item"]) > 0 {
				log.Println("strings.ToLower(g.Patrons[match].Username) == \"garcon\"")
				itemAdded = true
			}
		}
	}

	if g.debug {
		log.Printf("Checked if message received was an order command. Returning %v\n", itemAdded)
	}
	return
}

// NewGarcon constructs a new instance of Garcon and establishes all the behavior functions
// Note that while it is possible for us to make this code broader and more reusable, I don't
// have an intense interest in doing that right now, and I think such a structure would render
// the code either hilariously unreadable, complicated, and most likely both
func NewGarcon() *Garcon {
	g := &Garcon{}
	g.Stage = "uninitiated"

	// possible returns: affirmative, negative, additive, cancelling, irrelevant, indeterminable
	g.MessageTypeFuncs = map[string]func(slack.Msg) (string, error){
		"uninitiated": func(m slack.Msg) (string, error) {
			if g.CancellationCommandIssued(m.Text) {
				return "cancelling", nil
			}
			if strings.ToLower(m.Text) == "oh, garçon?" || strings.ToLower(m.Text) == "oh, @garcon?" {
				return "affirmative", nil
			}
			return "irrelevant", nil
		},
		"prompted": func(m slack.Msg) (string, error) {
			if g.CancellationCommandIssued(m.Text) {
				return "cancelling", nil
			}
			negative := responseIsNegative(m.Text)
			if negative || m.User != g.InterlocutorID {
				return "negative", nil
			}
			match, err := findElementsInString(orderInitiationPattern, []string{"restaurant"}, m.Text)
			restaurant := match["restaurant"]
			if err != nil {
				return "indeterminable", err
			}
			if len(restaurant) > 0 {
				return "affirmative", nil
			}
			return "indeterminable", nil
		},
		"ordering": func(m slack.Msg) (string, error) {
			if g.CancellationCommandIssued(m.Text) {
				return "cancelling", nil
			}
			if g.ItemAddedToOrder(m.Text) {
				return "affirmative", nil
			}
			return "indeterminable", nil
		},
		"confirmation": func(m slack.Msg) (string, error) {
			if g.CancellationCommandIssued(m.Text) {
				return "cancelling", nil
			}
			return "indeterminable", nil
		},
	}

	genericCancelReponse := func(m slack.Msg) []slack.OutgoingMessage {
		g.Reset()
		t := "Very well then, I'll disappear for now!"
		return []slack.OutgoingMessage{slack.OutgoingMessage{Channel: m.Channel, Text: t}}
	}

	genericHelpResponse := func(m slack.Msg, examples []string) []slack.OutgoingMessage {
		s := "\n • "
		t := fmt.Sprintf("I'm sorry, @%v, I couldn't understand what you said. Here are some things I might understand:%v%v\n", g.Patrons[m.User].Name, s, strings.Join(examples, s))
		return []slack.OutgoingMessage{slack.OutgoingMessage{Channel: m.Channel, Text: t}}
	}

	g.ReactionFuncs = map[string]map[string]func(slack.Msg) []slack.OutgoingMessage{
		"uninitiated": map[string]func(m slack.Msg) []slack.OutgoingMessage{
			"affirmative": func(m slack.Msg) []slack.OutgoingMessage {
				t := fmt.Sprintf("Hi, @%v! Would you like to place an order?", g.Patrons[m.User].Name)
				g.InterlocutorID = m.User
				g.Stage = "prompted"

				return []slack.OutgoingMessage{
					slack.OutgoingMessage{Channel: "C1N3MEUMN", Text: t},
				}
			},
		},
		"prompted": map[string]func(m slack.Msg) []slack.OutgoingMessage{
			"affirmative": func(m slack.Msg) []slack.OutgoingMessage {
				exampleResponses := []string{
					"We'd like to place an order from the Chili's at 45th & Lamar",
					"We would like to order from the Chili's at 45th & Lamar",
				}
				match, err := findElementsInString(orderInitiationPattern, []string{"restaurant"}, m.Text)
				restaurant := match["restaurant"]

				if err != nil || len(restaurant) == 0 {
					return genericHelpResponse(m, exampleResponses)
				}
				t := fmt.Sprintf("Okay, what would everyone like from %v?", restaurant)
				g.Stage = "ordering"
				return []slack.OutgoingMessage{slack.OutgoingMessage{Channel: m.Channel, Text: t}}
			},
			"cancelling": genericCancelReponse,
		},
		"ordering": map[string]func(m slack.Msg) []slack.OutgoingMessage{
			"affirmative": func(m slack.Msg) []slack.OutgoingMessage {
				t := fmt.Sprintf("Okay %v, I've got your order.", g.Patrons[m.User].Name)
				return []slack.OutgoingMessage{slack.OutgoingMessage{Channel: m.Channel, Text: t}}
			},
			"cancelling": genericCancelReponse,
		},
		"confirmation": map[string]func(m slack.Msg) []slack.OutgoingMessage{
			"cancelling": genericCancelReponse,
		},
	}

	return g
}
