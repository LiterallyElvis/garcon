package main

import (
	"github.com/literallyelvis/slacker"
	"log"
	"os"
)

func main() {
	log.Printf("starting Garcon with the following key: %v\n", os.Getenv("GARCON_TOKEN"))
	sb, err := slacker.NewBot(os.Getenv("GARCON_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}
	sb.Debug = true

	g := NewGarcon()
	g.debug = true
	g.Patrons = sb.GetSlackUsers()

	for {
		// read each incoming message
		m, err := sb.GetMessage()
		if err != nil {
			log.Fatal(err)
		}

		if m.Type == "message" {
			mt, err := g.MessageTypeFuncs[g.Stage](m)
			if err != nil {
				log.Printf("error determining message type: %v", err)
			}
			if _, ok := g.ReactionFuncs[g.Stage][mt]; ok {
				responses := g.ReactionFuncs[g.Stage][mt](m)
				for _, response := range responses {
					if len(response.Text) > 0 {
						// log.Printf("Sending this response to channel %v: %v", m.Channel, nm.Text)
						err := sb.SendMessage(response)
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
	}
}
