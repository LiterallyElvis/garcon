package main

import (
	"fmt"
	"github.com/literallyelvis/slacker"
	"log"
	"os"
	"strings"
)

type Order struct {
	Begun                bool
	Stage                string
	RequestedOrders      map[string]string
	RequestedRestauraunt string
	ActualRestaurant     string
}

var patterns map[string][]string

func init() {
	patterns = map[string][]string{
		"first": []string{
			"(oh, garçon?)",
		},
		"second": []string{
			"(we'd|we would) (like to) (place an)? ?(order) (for|from)? ?(?P<restaurant>.*)",
		},
	}
}

func main() {
	// start a websocket-based Real Time API session
	sb, err := slacker.NewBot(os.Getenv("GARCON_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Bot ready, ^C exits")
	users := sb.GetSlackUsers()
	order := Order{}

	for {
		// read each incoming message
		m, err := sb.GetMessage()
		if err != nil {
			log.Fatal(err)
		}

		if m.Type == "message" {
			// sb.logMessageObj(m)
			shouldRespond := false
			var nm slacker.SlackMessage
			if strings.ToLower(m.Text) == "oh, garçon?" && !order.Begun {
				shouldRespond = true
				nm = slacker.SlackMessage{Channel: m.Channel, Text: fmt.Sprintf("Hi, @%v! Would you like to place an order?", users[m.User].Username)}
				order.Stage = "prompted"
			} else if cleanString(m.Text) == "we'd like to order from (.*)" && order.Stage == "prompted" {
				shouldRespond = true
				restauraunt := "Chili's"              // parseRestauraunt
				suggestions := []string{"Applebee's"} // checkForRestauraunt("")

				var response string
				if len(suggestions) > 0 {
					suggestionsString := ""
					for i, s := range suggestions {
						st := fmt.Sprintf("\n%v: %v", i, s)
						suggestionsString = fmt.Sprintf("%v%v", suggestionsString, st)
					}
					response = fmt.Sprintf("I'm sorry, I don't know what that is. Did you mean one of these?\n\n%v", suggestionsString)
				} else {
					response = fmt.Sprintf("Okay, what would everybody like from %v?", restauraunt)
				}

				nm = slacker.SlackMessage{Channel: m.Channel, Text: response}
				order.Stage = "restauraunt established"
			}

			if shouldRespond && len(nm.Text) > 0 {
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
