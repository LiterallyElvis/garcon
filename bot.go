package main

import (
	"github.com/literallyelvis/slacker"
	"log"
	"os"
)

func main() {
	sb, err := slacker.NewBot(os.Getenv("GARCON_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	g := NewGarcon()
	g.Patrons = sb.GetSlackUsers()

	for {
		// read each incoming message
		m, err := sb.GetMessage()
		if err != nil {
			log.Fatal(err)
		}

		if m.Type == "message" {
			var nm slacker.SlackMessage

			mt, err := g.MessageTypeFuncs[g.Stage](m)
			if err != nil {
				log.Printf("error determining message type: %v", err)
			}
			response := g.ReactionFuncs[g.Stage][mt](m)
			if len(response.Text) > 0 {
				// log.Printf("Sending this response to channel %v: %v", m.Channel, nm.Text)
				err = sb.SendMessage(nm)
				if err != nil {
					log.Printf("error sending message: %v", err)
				} else {
					// log.Println("Message successfully sent!")
				}
			}
		}
	}
}
