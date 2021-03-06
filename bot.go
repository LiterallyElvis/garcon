package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jasonmoo/ghostmates"
	"github.com/nlopes/slack"
	"googlemaps.github.io/maps"
)

var g *Garcon
var sb *slack.Client
var rtm *slack.RTM
var allowedChannels []string
var errorEncounteredDoingSetup bool

func init() {
	allowedChannels = []string{"food", "bot-tester", "garcon_test"}
	sb = slack.New(os.Getenv("GARCON_TOKEN"))
	rtm = sb.NewRTM()

	g = NewGarcon()
	g.SelfName = "garcon"
	g.debug = true

	users, err := sb.GetUsers()
	if err != nil {
		errorEncounteredDoingSetup = true
		if fmt.Sprintf("%v", err) == "Post https://slack.com/api/users.list: dial tcp: lookup slack.com: no such host" {
			log.Printf("No internet connectivity, skipping setup. :(")
			return
		}
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

	customerID := os.Getenv("POSTMATES_CUSTOMER_ID")
	sandboxKey := os.Getenv("POSTMATES_API_TOKEN")
	duration, err := time.ParseDuration("1m")
	if err != nil {
		log.Fatal(err)
	}
	g.PostmatesClient = ghostmates.NewClient(customerID, sandboxKey, duration)
	g.OrderDestination = ghostmates.NewDeliverySpot(os.Getenv("GARCON_DESTINATION_NAME"), os.Getenv("GARCON_DESTINATION_ADDRESS"), os.Getenv("GARCON_DESTINATION_NUMBER"))

	apiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	c, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("fatal error establishing client: %s", err)
	}
	g.GoogleMapsClient = c
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
	for _, response := range responses {
		if len(response.Text) > 0 && sliceContainsString(response.Channel, g.AllowedChannels) {
			rtm.SendMessage(rtm.NewOutgoingMessage(response.Text, response.Channel))
		} else {
			log.Printf("I couldn't send this message\n\t%v\n", response.Text)
		}
	}
}

func main() {
	if !errorEncounteredDoingSetup {
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
}
