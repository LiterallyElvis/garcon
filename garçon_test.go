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
	m.Text = "<@GARCONBOT> you can place our order now."

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
