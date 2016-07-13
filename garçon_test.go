package main

import (
	// "github.com/literallyelvis/solid"
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
	"testing"
)

func returnGarconAndEmptyMessage() (*Garcon, slack.Msg) {
	dummyUserID := "LOLWTFBBQ"
	g := NewGarcon()
	g.SelfID = "GARCONBOT"
	g.Patrons = map[string]slack.User{
		"LOLWTFBBQ": slack.User{
			ID:   dummyUserID,
			Name: "brainfart",
		},
		"GARCONBOT": slack.User{
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

// orderConfirmationRequestPattern = "(<@(?P<user>\\w+)>(:|,)?(\\s+)you can place our order now(.*))"
// altOrderConfirmationPattern     = "((ok)?( |, )?<@(?P<user>\\w+)>(:|,)?(\\s+)I think (we are|we're) ready"

func TestPossibleValidCommands(t *testing.T) {
	patternsAndCommands := map[string][]string{
		"(we'd|we would) (like to) (place an)? ?(order) (for|from)? ?(?P<restaurant>.*)": []string{
			"We'd like to place an order from the Chili's on 45th & Lamar?",
			"We would like to place an order for the Chili's on 45th & Lamar?",
		},
		"(<@(?P<user>\\w+)>(:|,)?(\\s+)(abort|go away|leave|shut up))": []string{
			"<@GARCON>: abort",
			"<@GARCON>: go away",
			"<@GARCON>   LEAVE",
			"<@GARCON>, shut up",
		},
		"(<@(?P<user>\\w+)>(:|,)?(\\s+)((I would|I'd) like|I'll have) (?P<item>.*))": []string{
			"<@GARCON>: I would like the peach melba",
			"<@GARCON>:    I'd like the peach melba",
			"<@GARCON>: I'll have the peach melba",
		},
		"(<@(?P<user>\\w+)>(:|,)?(\\s+)(what does|what's) our order look like( so far)??)": []string{
			"<@GARCON>, what does our order look like?",
			"<@GARCON>: what's our order look like?",
			"<@GARCON>, what does our order look like so far?",
			"<@GARCON>: what's our order look like so far?",
		},
		"(ok)?( |, )?<@(?P<user>\\w+)>(:|,)?(\\s+)I think (we are|we're) ready( now)?": []string{
			"ok, <@GARCON>, I think we're ready",
			"ok, <@GARCON>: I think we're ready now",
			"ok, <@GARCON>   I think we are ready",
		},
	}

	for pattern, commands := range patternsAndCommands {
		for _, command := range commands {
			if !stringFitsPattern(pattern, command) {
				t.Errorf("%v didn't fit the pattern %v", command, pattern)
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
	assert.Equal(t, "I'm sorry, @brainfart, I couldn't understand what you said. Here are some things I might understand:\n • We'd like to place an order from the Chili's at 45th & Lamar\n • We would like to order from the Chili's at 45th & Lamar\n • @garcon, go away\n", messages[0].Text)
}

func TestGarconRespondsToOrderRequest(t *testing.T) {
	g, m := returnGarconAndEmptyMessage()
	g.Stage = "ordering"
	m.Text = "<@GARCONBOT> I'll have a peach melba"

	messages := g.RespondToMessage(m)
	// Garçon shouldn't respond to an order.
	assert.Equal(t, 0, len(messages))
}

func TestGarconRespondsToOrderConfirmationRequest(t *testing.T) {
	g, m := returnGarconAndEmptyMessage()

	g.Stage = "ordering"
	m.Text = "<@GARCONBOT> I'll have a peach melba"
	_ = g.RespondToMessage(m)

	_, m = returnGarconAndEmptyMessage()
	g.Stage = "ordering"
	m.Text = "ok, <@GARCON>: I think we're ready now"

	messages := g.RespondToMessage(m)
	assert.Equal(t, 3, len(messages))
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
