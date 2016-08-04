package main

import (
	"fmt"
	"github.com/nlopes/slack"
	"log"
	"os"
)

var g *Garcon
var sb *slack.Client
var rtm *slack.RTM
var allowedChannels []string

func init() {
	allowedChannels = []string{"garcon_test", "food", "bot-tester"}
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

func logMessage(m slack.Msg) {
	messageString := `
		Channel: %v
		User:    %v
		Text:    %v
	`
	log.Printf(messageString, m.Channel, m.User, m.Text)
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
			log.Printf("didn't send message because:\nlen(response.Text): %v\nresponse.Channel: %v\ng.AllowedChannels: %v\n", len(response.Text), response.Channel, g.AllowedChannels)
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
				fmt.Printf("Error: %s\n", ev.Error())
			}
		}
	}
}
