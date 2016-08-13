package main

import (
	"log"
	"os"

	"github.com/nlopes/slack"
)

var g *Garcon
var sb *slack.Client
var rtm *slack.RTM
var allowedChannels []string

func init() {
	allowedChannels = []string{"food", "bot-tester"}
	sb = slack.New(os.Getenv("GARCON_TOKEN"))
	rtm = sb.NewRTM()
	// sb.SetDebug(true)

	g = NewGarcon()
	g.SelfName = "garcon"
	g.debug = true

	users, err := sb.GetUsers()
	if err != nil {
		log.Printf("Error retrieving users:\n%v\n", err)
	}
	g.Patrons = make(map[string]slack.User)
	for _, u := range users {
		g.Patrons[u.ID] = u
	}
	g.FindBotSlackID()

	channels, err := sb.GetChannels(true)
	if err != nil {
		log.Fatal(err)
	}
	for _, ch := range channels {
		if sliceContainsString(ch.Name, allowedChannels) {
			g.AllowedChannels = append(g.AllowedChannels, ch.ID)
		}
	}
}

func makeIDToUserMap(in []slack.User) map[string]slack.User {
	users := make(map[string]slack.User)
	for _, u := range in {
		users[u.ID] = u
	}
	return users
}

func handleMessage(m slack.Msg) {
	responses := g.RespondToMessage(m)
	if g.debug {
		g.logGarconInfo(m)
	}
	for _, response := range responses {
		if len(response.Text) > 0 && sliceContainsString(response.Channel, g.AllowedChannels) {
			rtm.SendMessage(rtm.NewOutgoingMessage(response.Text, response.Channel))
		} else {
			log.Printf("I couldn't send this message\n\t%v\n", response.Text)
		}
	}
}

func main() {
	go rtm.ManageConnection()
	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {

			case *slack.MessageEvent:
				handleMessage(ev.Msg)

			case *slack.RTMError:
				log.Printf("I encountered a Slack related error: %s\n", ev.Error())
			}
		}
	}
}
