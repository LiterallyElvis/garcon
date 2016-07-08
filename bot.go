package main

import (
	"fmt"
	"github.com/literallyelvis/solid"
	"github.com/nlopes/slack"
	"log"
	"os"
)

var g *Garcon
var sb *slack.Client
var rtm *slack.RTM

func init() {
	sb = slack.New(os.Getenv("GARCON_TOKEN"))
	rtm = sb.NewRTM()
	// sb.SetDebug(true)

	g = NewGarcon()
	g.debug = true

	users, err := sb.GetUsers()
	if err != nil {
		log.Printf("Error retrieving users:\n%v\n", err)
	}
	g.Patrons = makeIDToUserMap(users)
	g.FindBotSlackID()
	c, err := solid.New(os.Getenv("FAVOR_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}
	g.FavorClient = c
}

func makeIDToUserMap(in []slack.User) map[string]slack.User {
	users := make(map[string]slack.User)
	for _, u := range in {
		users[u.ID] = u
	}
	return users
}

func handleMessage(m slack.Msg) {
	if g.debug {
		status := `
			Stage:          %v
			InterlocutorID: %v

		`
		log.Printf(status, g.Stage, g.InterlocutorID)
		messageString := `
			Channel: %v
			User:    %v
			Text:    %v

		`
		log.Printf(messageString, m.Channel, m.User, m.Text)
	}
	if m.User == g.SelfID || len(m.User) == 0 {
		return
	}

	mt, err := g.MessageTypeFuncs[g.Stage](m)
	if err != nil {
		log.Printf("error determining message type: %v", err)
	}
	if g.debug {
		log.Printf("determined message type to be %v\n", mt)
	}

	if _, ok := g.ReactionFuncs[g.Stage][mt]; ok {
		responses := g.ReactionFuncs[g.Stage][mt](m)
		for _, response := range responses {
			if len(response.Text) > 0 {
				rtm.SendMessage(rtm.NewOutgoingMessage(response.Text, response.Channel))
				if err != nil {
					log.Printf("error sending message:\n%v\n", err)
				} else {
					log.Printf("Successfully sent the following message:\n\t%v\n", response.Text)
				}
			}
		}
	} else {
		if g.debug {
			log.Printf("No reaction functions found for current:\n\tStage: %v\n\tMessageType: %v\n", g.Stage, mt)
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
