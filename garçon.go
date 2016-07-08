package main

import (
	"fmt"
	"github.com/literallyelvis/solid"
	"github.com/nlopes/slack"
	"log"
	"strings"
)

const (
	orderInitiationPattern          = "(we'd|we would) (like to) (place an)? ?(order) (for|from)? ?(?P<restaurant>.*)"
	abortCommandPattern             = "(<@(?P<user>\\w+)>(:|,)?(\\s+)(abort|go away|leave|shut up))"
	orderPlacingPattern             = "(<@(?P<user>\\w+)>(:|,)?(\\s+)((I would|I'd) like|I'll have) (?P<item>.*))"
	orderStatusRequestPattern       = "(<@(?P<user>\\w+)>(:|,)?(\\s+)(what does|what's) our order look like( so far)??)"
	orderConfirmationRequestPattern = "(<@(?P<user>\\w+)>(:|,)?(\\s+)you can place our order now(.*))"
	altOrderConfirmationPattern     = "(<@(?P<user>\\w+)>(:|,)?(\\s+)I think (we are|we're) ready"
)

// Garcon is our order taking bot! ヽ(゜∇゜)ノ
type Garcon struct {
	debug                bool
	SelfName             string
	SelfID               string
	Stage                string
	AllowedChannels      []string
	InterlocutorID       string
	RequestedRestauraunt string
	FavorClient          *solid.Client
	FavorOrder           solid.Favor
	Order                map[string]string
	Patrons              map[string]slack.User
	MessageTypeFuncs     map[string]func(slack.Msg) (string, error)
	ReactionFuncs        map[string]map[string]func(slack.Msg) []slack.OutgoingMessage
	CommandExamples      map[string][]string
}

// FindBotSlackID TODO: document
func (g *Garcon) FindBotSlackID() {
	for id, p := range g.Patrons {
		if strings.ToLower(p.Name) == strings.ToLower(g.SelfName) {
			g.SelfID = id
		}
	}
}

// Reset wipes the state of Garcon
func (g *Garcon) Reset() {
	g.Stage = "uninitiated"
	g.Order = make(map[string]string)
}

// CancellationCommandIssued returns whether or not the most recent command
// is a show-stopping cancellation command directed at Garcon
func (g Garcon) cancellationCommandIssued(m string) (abortCommandIssued bool) {
	if stringFitsPattern(abortCommandPattern, m) {
		match, _ := findElementsInString(abortCommandPattern, []string{"user"}, m)
		user := match["user"]
		if _, ok := g.Patrons[user]; ok {
			if strings.ToLower(g.Patrons[user].Name) == "garcon" {
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
func (g Garcon) itemAddedToOrder(m string) (itemAdded bool) {
	if stringFitsPattern(orderPlacingPattern, m) {
		matches, _ := findElementsInString(orderPlacingPattern, []string{"user", "item"}, m)
		user := matches["user"]
		if _, ok := g.Patrons[user]; ok {
			if strings.ToLower(g.Patrons[user].Name) == "garcon" && len(matches["item"]) > 0 {
				itemAdded = true
			}
		}
	}

	if g.debug {
		log.Printf("Checked if message received was to add an item to the order. Returning %v\n", itemAdded)
	}
	return
}

// OrderStatusCheckRequested TODO: Document
func (g Garcon) orderStatusCheckRequested(m string) (requested bool) {
	if stringFitsPattern(orderStatusRequestPattern, m) {
		requested = true
	}

	if g.debug {
		log.Printf("Checked if message received was an order status reequest. Returning %v\n", requested)
	}
	return
}

// ReadyToPlaceOrder TODO: Document
func (g Garcon) readyToPlaceOrder(m string) (ready bool) {
	if stringFitsPattern(orderConfirmationRequestPattern, m) {
		ready = true
	}

	if g.debug {
		log.Printf("Checked if message received was an order status reequest. Returning %v\n", ready)
	}
	return
}

// HelpRequested TODO: Document
func (g Garcon) helpRequested(m string) (helpRequested bool) {
	return strings.ToLower(m) == "@garcon, help me!"
}

func (g Garcon) suggestHelpCommandResponse(m slack.Msg) []slack.OutgoingMessage {
	t := fmt.Sprintf("I'm sorry, @%v, I couldn't understand what you said. For help, say \"@garcon, help me!\"", g.Patrons[m.User].Name)
	return []slack.OutgoingMessage{slack.OutgoingMessage{Channel: m.Channel, Text: t}}
}

func (g Garcon) genericHelpResponse(m slack.Msg) []slack.OutgoingMessage {
	s := "\n • "
	examples := append(g.CommandExamples[g.Stage], g.CommandExamples["always"]...)
	t := fmt.Sprintf("I'm sorry, @%v, I couldn't understand what you said. Here are some things I might understand:%v%v\n", g.Patrons[m.User].Name, s, strings.Join(examples, s))
	return []slack.OutgoingMessage{slack.OutgoingMessage{Channel: m.Channel, Text: t}}
}

func (g Garcon) orderStatusResponse(m slack.Msg) []slack.OutgoingMessage {
	// TODO: Make this a more generic function
	orders := ""
	for u, o := range g.Order {
		orders = fmt.Sprintf("%v%v: %v\n", orders, strings.Title(u), o)
	}
	orders = strings.TrimSpace(orders)
	statusTemplate := "Here's what I have for your order from %v:\n```\n%v\n```"
	statusMessage := fmt.Sprintf(statusTemplate, g.RequestedRestauraunt, orders)
	return []slack.OutgoingMessage{slack.OutgoingMessage{Channel: m.Channel, Text: statusMessage}}
}

func (g Garcon) genericCancelReponse(m slack.Msg) []slack.OutgoingMessage {
	g.Reset()
	t := "Very well then, I'll disappear for now!"
	return []slack.OutgoingMessage{slack.OutgoingMessage{Channel: m.Channel, Text: t}}
}

// NewGarcon constructs a new instance of Garcon and establishes all the behavior functions
// Note that while it is possible for us to make this code broader and more reusable, I don't
// have an intense interest in doing that right now, and I think such a structure would render
// the code either hilariously unreadable, complicated, and most likely both
func NewGarcon() *Garcon {
	g := &Garcon{
		SelfName: "garcon",
		AllowedChannels: []string{
			"garcon_test",
			"food",
		},
		Stage: "uninitiated",
		Order: make(map[string]string),
	}

	g.CommandExamples = map[string][]string{
		"uninitiated": []string{
			"oh, garçon?",
		},
		"prompted": []string{
			"We'd like to place an order from the Chili's at 45th & Lamar",
			"We would like to order from the Chili's at 45th & Lamar",
		},
		"ordering": []string{
			"@garcon, I'd like a banana",
			"@garcon I'll have the tuna melt",
			"@garcon, what's our order look like so far?",
		},
		"confirmation": []string{
			"yes",
			"no",
		},
		"always": []string{
			"@garcon, what's our order look like so far?",
			"@garcon, go away",
		},
	}

	// possible returns: affirmative, negative, additive, cancelling, irrelevant, indeterminable
	g.MessageTypeFuncs = map[string]func(slack.Msg) (string, error){
		"uninitiated": func(m slack.Msg) (string, error) {
			if g.cancellationCommandIssued(m.Text) {
				return "cancelling", nil
			}
			if strings.ToLower(m.Text) == "oh, garçon?" || strings.ToLower(m.Text) == "oh, @garcon?" {
				return "affirmative", nil
			}
			return "irrelevant", nil
		},
		"prompted": func(m slack.Msg) (string, error) {
			if g.cancellationCommandIssued(m.Text) {
				return "cancelling", nil
			}
			if g.helpRequested(m.Text) {
				return "inquisitive", nil
			}
			negative := responseIsNegative(m.Text)
			if negative || m.User != g.InterlocutorID {
				return "negative", nil
			}
			if stringFitsPattern(orderInitiationPattern, m.Text) {
				return "affirmative", nil
			}
			return "indeterminable", nil
		},
		"ordering": func(m slack.Msg) (string, error) {
			if g.cancellationCommandIssued(m.Text) {
				return "cancelling", nil
			}
			if g.helpRequested(m.Text) {
				return "inquisitive", nil
			}
			if g.itemAddedToOrder(m.Text) {
				return "contributing", nil
			}
			if g.orderStatusCheckRequested(m.Text) {
				return "status", nil
			}
			if g.readyToPlaceOrder(m.Text) {
				return "affirmative", nil
			}
			return "indeterminable", nil
		},
		"confirmation": func(m slack.Msg) (string, error) {
			if g.cancellationCommandIssued(m.Text) {
				return "cancelling", nil
			}
			if g.helpRequested(m.Text) {
				return "inquisitive", nil
			}
			if responseIsAffirmative(m.Text) {
				return "affirmative", nil
			}
			if responseIsNegative(m.Text) {
				return "negative", nil
			}
			return "indeterminable", nil
		},
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
				match, err := findElementsInString(orderInitiationPattern, []string{"restaurant"}, m.Text)
				restaurant := match["restaurant"]

				if err != nil || len(restaurant) == 0 {
					return g.suggestHelpCommandResponse(m)
				}
				t := fmt.Sprintf("Okay, what would everyone like from %v?", restaurant)
				g.Stage = "ordering"
				g.RequestedRestauraunt = restaurant

				return []slack.OutgoingMessage{slack.OutgoingMessage{Channel: m.Channel, Text: t}}
			},
			"inquisitive": g.genericHelpResponse,
			"cancelling":  g.genericCancelReponse,
		},
		"ordering": map[string]func(m slack.Msg) []slack.OutgoingMessage{
			"affirmative": func(m slack.Msg) []slack.OutgoingMessage {
				g.Stage = "confirmation"

				return []slack.OutgoingMessage{
					slack.OutgoingMessage{Channel: m.Channel, Text: "Alright, then"},
					g.orderStatusResponse(m)[0],
					slack.OutgoingMessage{Channel: m.Channel, Text: "Is that correct?"},
				}
			},
			"contributing": func(m slack.Msg) []slack.OutgoingMessage {
				matches, err := findElementsInString(orderPlacingPattern, []string{"item"}, m.Text)
				item := matches["item"]

				if err != nil || len(item) == 0 {
					return g.genericHelpResponse(m)
				}

				g.Order[g.Patrons[m.User].Name] = item
				// t := fmt.Sprintf("Okay @%v, I've got your order.", g.Patrons[m.User].Name)
				// return []slack.OutgoingMessage{slack.OutgoingMessage{Channel: m.Channel, Text: t}}
				return []slack.OutgoingMessage{slack.OutgoingMessage{}}
			},
			"inquisitive": g.genericHelpResponse,
			"cancelling":  g.genericCancelReponse,
			"status":      g.orderStatusResponse,
		},
		"confirmation": map[string]func(m slack.Msg) []slack.OutgoingMessage{
			"affirmative": func(m slack.Msg) []slack.OutgoingMessage {
				t := "Okay, I'll send this order off!"
				return []slack.OutgoingMessage{slack.OutgoingMessage{Channel: m.Channel, Text: t}}
			},
			"cancelling":     g.genericCancelReponse,
			"inquisitive":    g.genericHelpResponse,
			"indeterminable": g.genericHelpResponse,
		},
	}

	return g
}
