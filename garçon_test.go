package main

import (
	"fmt"
	"testing"

	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
)

func returnGarconAndEmptyMessage() (*Garcon, slack.Msg) {
	dummyUserID := "LOLWTFBBQ"
	g := NewGarcon()
	g.SelfID = "G4RC0NB0T"
	g.debug = true
	g.Patrons = map[string]slack.User{
		"LOLWTFBBQ": slack.User{
			ID:   dummyUserID,
			Name: "brainfart",
		},
		"G4RC0NB0T": slack.User{
			ID:   g.SelfID,
			Name: "garcon",
		},
	}
	g.InterlocutorID = dummyUserID
	m := slack.Msg{
		User:    dummyUserID,
		Channel: "whocares",
	}
	return g, m
}

func TestGarconConstructsWithoutError(t *testing.T) {
	g := NewGarcon()
	g.Stage = "test"
}

func TestGarconReset(t *testing.T) {
	expected := Garcon{
		Stage:                "uninitiated",
		InterlocutorID:       "",
		RequestedRestauraunt: "",
		Order:                map[string]string{},
	}

	actual := Garcon{
		Stage:                "whatever",
		InterlocutorID:       "who cares",
		RequestedRestauraunt: "Greasy Gus's Frog Emporium",
		Order: map[string]string{
			"Gary": "frog balls",
		},
	}
	actual.Reset()

	assert.Equal(t, expected, actual)
}

func TestPossibleValidCommands(t *testing.T) {
	patternsAndCommands := map[string][]string{
		"(we'd|we would) (like to) (place an)? ?(order) (for|from)? ?(?P<restaurant>.*)": []string{
			"We'd like to place an order from the Chili's on 45th & Lamar",
			"We would like to place an order for the Chili's on 45th & Lamar",
			"We'd like to order from the Chili's on 45th & Lamar",
			"WE WOULD LIKE TO ORDER FROM THE CHILI'S ON 45TH AND LAMAR",
		},
		"(<@(?P<user>[0-9A-Z]{9})>(:|,)?(\\s*?)(abort|go away|leave|shut up))": []string{
			"<@G4RC0NB0T>: abort",
			"<@G4RC0NB0T>: go away",
			"<@G4RC0NB0T>   LEAVE",
			"<@G4RC0NB0T>, shut up",
		},
		"(<@(?P<user>[0-9A-Z]{9})>(:|,)?(\\s*?)(help|help me|help us)(!)?)": []string{
			"<@G4RC0NB0T>: help me",
			"<@G4RC0NB0T>: help!",
			"<@G4RC0NB0T>   help us!",
			"<@G4RC0NB0T>, help us",
		},
		"(<@(?P<user>[0-9A-Z]{9})>(:|,)?(\\s*?)((I would|I'd) like|I'll have) (?P<item>.*))": []string{
			"<@G4RC0NB0T>: I would like the peach melba",
			"<@G4RC0NB0T>:    I'd like the peach melba",
			"<@G4RC0NB0T> I'll have the peach melba",
			"<@G4RC0NB0T>, I'll have the poutine",
			"<@G4RC0NB0T>:I’ll have a Super Bol",
		},
		"(<@(?P<user>[0-9A-Z]{9})>(:|,)?(\\s*?)(what does|what's) our order look like( so far)??)": []string{
			"<@G4RC0NB0T>, what does our order look like?",
			"<@G4RC0NB0T>: what's our order look like?",
			"<@G4RC0NB0T>, what does our order look like so far?",
			"<@G4RC0NB0T>: what's our order look like so far?",
		},
		"(ok)?( |, )?<@(?P<user>[0-9A-Z]{9})>(:|,)?(\\s*?)I think (we are|we're) ready( now)?": []string{
			"ok, <@G4RC0NB0T>, I think we're ready",
			"ok, <@G4RC0NB0T>: I think we're ready now",
			"ok, <@G4RC0NB0T>   I think we are ready",
		},
	}

	for pattern, commands := range patternsAndCommands {
		for _, command := range commands {
			if !stringFitsPattern(pattern, command) {
				t.Errorf("\n\t'%v'\ndidn't fit the pattern\n\t'%v'", command, pattern)
			}
		}
	}
}

func TestGarconRespondsToHello(t *testing.T) {
	g, m := returnGarconAndEmptyMessage()
	m.Text = "oh, garçon?"
	messages := g.RespondToMessage(m)
	assert.Equal(t, 1, len(messages))
	assert.Equal(t, "Hi, @brainfart! Would you like to place an order?", messages[0].Text)
}

func TestGarconUnderstandsWhenAddressed(t *testing.T) {
	g, _ := returnGarconAndEmptyMessage()
	validMessages := []slack.Msg{
		slack.Msg{Text: "ok <@G4RC0NB0T>, "},
		slack.Msg{Text: "okay <@G4RC0NB0T>, "},
		slack.Msg{Text: "ok, <@G4RC0NB0T>, "},
		slack.Msg{Text: "okay, <@G4RC0NB0T>, "},

		slack.Msg{Text: "ok <@G4RC0NB0T>: "},
		slack.Msg{Text: "okay <@G4RC0NB0T>: "},
		slack.Msg{Text: "ok, <@G4RC0NB0T>: "},
		slack.Msg{Text: "okay, <@G4RC0NB0T>: "},

		slack.Msg{Text: "ok <@G4RC0NB0T>,"},
		slack.Msg{Text: "okay <@G4RC0NB0T>,"},
		slack.Msg{Text: "ok, <@G4RC0NB0T>:"},
		slack.Msg{Text: "okay, <@G4RC0NB0T>:"},

		slack.Msg{Text: "ok <@G4RC0NB0T>"},
		slack.Msg{Text: "okay <@G4RC0NB0T>"},
		slack.Msg{Text: "ok, <@G4RC0NB0T>"},
		slack.Msg{Text: "okay, <@G4RC0NB0T>"},
	}

	for _, m := range validMessages {
		assert.True(t, g.MessageAddressesGarcon(m), fmt.Sprintf("Message '%v' not identified as messaging Garcon", m.Text))
	}
}

func TestGarconDoesNotRespondWhenUninitiated(t *testing.T) {
	g, m := returnGarconAndEmptyMessage()
	m.Text = "Oh yeah...oh yeah...oh yeah...The moon...beautiful...the sun...even more beautiful"
	messages := g.RespondToMessage(m)
	assert.Equal(t, 0, len(messages))
}

func TestGarconRespondsToRestaurantRequest(t *testing.T) {
	g, m := returnGarconAndEmptyMessage()
	g.Stage = "prompted"
	m.Text = "We would like to order from the Chili's on 45th & Lamar"

	messages := g.RespondToMessage(m)
	assert.Equal(t, 1, len(messages))
	assert.Equal(t, "Okay, what would everyone like from the Chili's on 45th & Lamar?", messages[0].Text)
}

func TestGarconRespondsToNonInterlocutor(t *testing.T) {
	g, m := returnGarconAndEmptyMessage()
	g.InterlocutorID = "SOMEJERK"
	g.Patrons["SOMEJERK"] = slack.User{
		ID:   "SOMEJERK",
		Name: "whocares",
	}
	g.Stage = "prompted"
	m.Text = "We would like to order from the Chili's on 45th & Lamar"

	messages := g.RespondToMessage(m)
	assert.Equal(t, 0, len(messages))
}

func TestGarconRespondsToInvalidResponseAfterPrompt(t *testing.T) {
	g, m := returnGarconAndEmptyMessage()
	g.Stage = "prompted"
	m.Text = "I WANT A BIG OL' HEAP O' CHILI RIGHT NOW GOL DURNIT!"

	messages := g.RespondToMessage(m)
	assert.Equal(t, 1, len(messages))
	assert.Equal(t, "I'm sorry, @brainfart, I couldn't understand what you said. Here are some things I might understand:\n • We'd like to place an order for the Chili's at 45th & Lamar\n • We would like to order from the Chili's at 45th & Lamar\n • @garcon, go away\n • @garcon, help!\n", messages[0].Text)
}

func TestGarconRespondsToOrderRequest(t *testing.T) {
	g, m := returnGarconAndEmptyMessage()
	g.Stage = "ordering"
	m.Text = "<@G4RC0NB0T> I'll have a bananas foster"

	messages := g.RespondToMessage(m)
	assert.Equal(t, 1, len(messages))
	assert.Equal(t, "Okay @brainfart, I've got your order.", messages[0].Text)
}

func TestGarconRespondsToOrderConfirmationRequest(t *testing.T) {
	g, m := returnGarconAndEmptyMessage()

	g.Stage = "ordering"
	m.Text = "<@G4RC0NB0T> I'll have a peach melba"
	_ = g.RespondToMessage(m)

	m = slack.Msg{Text: "ok, <@G4RC0NB0T>: I think we're ready now"}

	messages := g.RespondToMessage(m)
	assert.Equal(t, 3, len(messages))
	if !assert.Equal(t, 3, len(messages)) {
		t.FailNow()
	}
	assert.Equal(t, "Alright, then!", messages[0].Text)
	assert.Equal(t, "Here's what I have for your order from :\n```\nBrainfart: a peach melba\n```", messages[1].Text)
	assert.Equal(t, "Is that correct?", messages[2].Text)
}

func TestGarconRespondsToOrderConfirmation(t *testing.T) {
	g, m := returnGarconAndEmptyMessage()
	g.Stage = "confirmation"
	m.Text = "yep!"

	messages := g.RespondToMessage(m)
	assert.Equal(t, 1, len(messages))
	assert.Equal(t, "Okay, I'll send this order off!", messages[0].Text)
}
