package main

import (
	"fmt"
	"log"
	"os"

	"github.com/nlopes/slack"
)

var g *Garcon
var sb *slack.Client

func init() {
	// log.Printf("starting Garcon with the following key: %v\n", os.Getenv("GARCON_TOKEN"))
	sb = slack.New(os.Getenv("GARCON_TOKEN"))
	logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)
	// sb.SetDebug(true)

	g = NewGarcon()

	users, err := sb.GetUsers()
	if err != nil {
		log.Printf("Error retrieving users:\n%v\n", err)
	}
	g.Patrons = makeIDToUserMap(users)
}

func makeIDToUserMap(in []slack.User) map[string]slack.User {
	users := make(map[string]slack.User)
	for _, u := range in {
		users[u.ID] = u
	}
	return users
}

func handleMessage(m slack.Msg) {
	mt, err := g.MessageTypeFuncs[g.Stage](m)
	if err != nil {
		log.Printf("error determining message type: %v", err)
	}
	if _, ok := g.ReactionFuncs[g.Stage][mt]; ok {
		responses := g.ReactionFuncs[g.Stage][mt](m)
		for _, response := range responses {
			if len(response.Text) > 0 {
				// log.Printf("Sending this response to channel %v: %v", m.Channel, nm.Text)
				// err := sb.SendMessage(response)
				sb.PostMessage(m.Channel, response.Text, slack.PostMessageParameters{})
				if err != nil {
					log.Printf("error sending message:\n%v\n", err)
				} else {
					log.Printf("Successfully sent the following message:\n\t%v\n", response.Text)
				}
			}
		}
	} else {
		if g.debug {
			log.Printf("No reaction functions found for current:\n\tstage: %v\n\tmessage type: %v\n", g.Stage, mt)
		}
	}
}

func main() {
	rtm := sb.NewRTM()
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