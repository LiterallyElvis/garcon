package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/jasonmoo/ghostmates"
	"github.com/nlopes/slack"
	"golang.org/x/net/context"
	"googlemaps.github.io/maps"
)

const (
	atGarconPattern                 = "((ok|okay)( |, )?)?<@(?P<user>[0-9A-Z]{9})>(:|,)?(\\s*?)"
	orderInitiationPattern          = "(we'd|we would) (like to) (place an)? ?(order) (for|from)? ?(?P<restaurant>.*)"
	abortCommandPattern             = "(abort|go away|leave|shut up)"
	helpRequestPattern              = "(help|help me|help us)(!)?"
	orderPlacingPattern             = "((I would|I'd) like|I'll have)(\\s*?)(?P<item>.*)"
	orderStatusRequestPattern       = "(what does|what's) our order look like( so far)??"
	orderConfirmationRequestPattern = "I think (we are|we're) ready( now)?"
)

// Garcon is our order taking bot! ヽ(゜∇゜)ノ
type Garcon struct {
	debug               bool
	SelfName            string
	SelfID              string
	Stage               string
	AllowedChannels     []string
	InterlocutorID      string
	RequestedRestaurant string
	ActualRestaurant    *maps.PlaceDetailsResult
	Order               map[string]string
	Patrons             map[string]slack.User
	MessageTypeFuncs    map[string]func(slack.Msg) (string, error)
	ReactionFuncs       map[string]map[string]func(slack.Msg) []slack.OutgoingMessage
	CommandExamples     map[string][]string

	PostmatesClient  *ghostmates.Client
	OrderDestination *ghostmates.DeliverySpot

	GoogleMapsClient *maps.Client
}

// FindBotSlackID iterates over all the slack users and figures out what
// Garcon's ID is
func (g *Garcon) FindBotSlackID() {
	for id, p := range g.Patrons {
		if strings.ToLower(p.Name) == strings.ToLower(g.SelfName) {
			g.SelfID = id
		}
	}
}

// MessageAddressesGarcon returns whether or not the message began with some variant
// of "ok, @garcon"
func (g *Garcon) MessageAddressesGarcon(m slack.Msg) bool {
	match, _ := findElementsInString(atGarconPattern, []string{"user"}, m.Text)
	user := match["user"]
	if _, ok := g.Patrons[user]; ok {
		if strings.ToLower(g.Patrons[user].Name) == "garcon" {
			return true
		}
	}
	return false
}

// Reset wipes the state of Garcon
func (g *Garcon) Reset() {
	g.Stage = "uninitiated"
	g.InterlocutorID = ""
	g.RequestedRestaurant = ""
	g.Order = make(map[string]string)
}

// RespondToMessage TODO: Document
func (g *Garcon) RespondToMessage(m slack.Msg) (responses []slack.OutgoingMessage) {
	if m.User == g.SelfID || len(m.User) == 0 {
		log.Printf("I've received a message, but I can't respond to it because something's up with the user who sent it: '%v'", m.User)
		return
	}

	if g.debug {
		log.Printf("I've received this message:\n\t'%v'\nand I'm going to try to determine what type it is\n", m.Text)
	}

	mt, err := g.MessageTypeFuncs[g.Stage](m)
	if err != nil && g.debug {
		log.Printf("I encountered this error determining the message type of that message:\n\t%v\n", err)
	}

	if g.debug {
		log.Printf("I've determined this message:\n\t%v\nto be %v", m.Text, mt)
	}

	if _, ok := g.ReactionFuncs[g.Stage][mt]; ok {
		responses = g.ReactionFuncs[g.Stage][mt](m)
		if g.debug {
			responseMessages := []string{}
			for _, r := range responses {
				responseMessages = append(responseMessages, r.Text)
			}
			debugResponses := strings.Join(responseMessages, "\n\t")
			log.Printf("I've decided to respond with the following responses:\n\t%v\n", debugResponses)
		}
	}

	return
}

// ItemAddedToOrder TODO: Document
func (g Garcon) itemAddedToOrder(m slack.Msg) (orderPlaced bool) {
	messageAddressesGarcon := g.MessageAddressesGarcon(m)
	messageIsPlacingAnOrder := stringFitsPattern(orderPlacingPattern, m.Text)
	if messageAddressesGarcon && messageIsPlacingAnOrder {
		orderPlaced = true
	} else {
		if g.debug {
			if messageAddressesGarcon {
				log.Printf("I think this message is meant for me: %v", m.Text)
			} else {
				log.Printf("This message wasn't meant for me: %v", m.Text)
			}
			if messageIsPlacingAnOrder {
				log.Printf("I think this message is trying to place an order: %v", m.Text)
			} else {
				log.Printf("I don't think this message is trying to place an order: %v", m.Text)
			}
		}
	}
	return
}

// TODO: Make these more generic
// CancellationCommandIssued returns whether or not the most recent command
// is a show-stopping cancellation command directed at Garcon
func (g Garcon) cancellationCommandIssued(m slack.Msg) bool {
	return g.MessageAddressesGarcon(m) && stringFitsPattern(abortCommandPattern, m.Text)
}

// OrderStatusCheckRequested TODO: Document
func (g Garcon) orderStatusCheckRequested(m slack.Msg) bool {
	return stringFitsPattern(orderStatusRequestPattern, m.Text)
}

// ReadyToPlaceOrder TODO: Document
func (g Garcon) readyToPlaceOrder(m slack.Msg) bool {
	return stringFitsPattern(orderConfirmationRequestPattern, m.Text)
}

// HelpRequested TODO: Document
func (g Garcon) helpRequested(m slack.Msg) bool {
	return g.MessageAddressesGarcon(m) && stringFitsPattern(helpRequestPattern, m.Text)
}

func (g *Garcon) suggestHelpCommandResponse(m slack.Msg) []slack.OutgoingMessage {
	t := fmt.Sprintf("I'm sorry, @%v, I couldn't understand what you said. For help, say \"@garcon, help me!\"", g.Patrons[m.User].Name)
	return []slack.OutgoingMessage{slack.OutgoingMessage{Channel: m.Channel, Text: t}}
}

func (g *Garcon) genericHelpResponse(m slack.Msg) []slack.OutgoingMessage {
	s := "\n • "
	examples := append(g.CommandExamples[g.Stage], g.CommandExamples["always"]...)
	t := fmt.Sprintf("I'm sorry, @%v, I couldn't understand what you said. Here are some things I might understand:%v%v\n", g.Patrons[m.User].Name, s, strings.Join(examples, s))
	return []slack.OutgoingMessage{slack.OutgoingMessage{Channel: m.Channel, Text: t}}
}

func (g *Garcon) helloGarcon(m slack.Msg) []slack.OutgoingMessage {
	t := fmt.Sprintf("Hi, @%v! Would you like to place an order?", g.Patrons[m.User].Name)
	g.InterlocutorID = m.User
	g.Stage = "prompted"

	return []slack.OutgoingMessage{
		slack.OutgoingMessage{Channel: m.Channel, Text: t},
	}
}

func (g *Garcon) validateRestaurant(m slack.Msg) []slack.OutgoingMessage {
	match, err := findElementsInString(orderInitiationPattern, []string{"restaurant"}, m.Text)
	g.RequestedRestaurant = match["restaurant"]

	if err != nil || len(g.RequestedRestaurant) == 0 {
		return g.suggestHelpCommandResponse(m)
	}

	gr := &maps.GeocodingRequest{Address: g.OrderDestination.Address}
	gresp, err := g.GoogleMapsClient.Geocode(context.TODO(), gr)
	if err != nil {
		log.Printf("fatal error geocoding Location: %s", err)
		return g.suggestHelpCommandResponse(m)
	}

	nsr := &maps.NearbySearchRequest{
		Location: &gresp[0].Geometry.Location,
		Radius:   10000, // In meters, this is completely arbitrary
		Keyword:  g.RequestedRestaurant,
		Name:     g.RequestedRestaurant,
		OpenNow:  true,
		MinPrice: "0",
		MaxPrice: "4",
		Type:     "restaurant",
	}

	places, err := g.GoogleMapsClient.NearbySearch(context.TODO(), nsr)
	log.Printf("I was able to find %v restaurants matching the requested name.", len(places.Results))
	if err != nil {
		log.Printf("fatal error conducting search: %s", err)
		em := "I'm sorry, I can't find a restaurant like that."
		return []slack.OutgoingMessage{slack.OutgoingMessage{Channel: m.Channel, Text: em}}
	}
	pdr := &maps.PlaceDetailsRequest{PlaceID: places.Results[0].PlaceID}
	place, err := g.GoogleMapsClient.PlaceDetails(context.TODO(), pdr)

	g.ActualRestaurant = &place
	log.Printf("I decided to go with the following restaurant:\n%v\n%v\n", g.ActualRestaurant.Name, g.ActualRestaurant.Vicinity)

	t := fmt.Sprintf("Okay, what would everyone like from %v?", g.ActualRestaurant.Name)
	g.Stage = "ordering"

	return []slack.OutgoingMessage{slack.OutgoingMessage{Channel: m.Channel, Text: t}}
}

func (g *Garcon) validateOrder(m slack.Msg) []slack.OutgoingMessage {
	g.Stage = "confirmation"

	if g.debug {
		log.Printf("I'm going to attempt to confirm the group order.")
	}

	return []slack.OutgoingMessage{
		slack.OutgoingMessage{Channel: m.Channel, Text: "Alright, then!"},
		g.orderStatusResponse(m)[0],
		slack.OutgoingMessage{Channel: m.Channel, Text: "Is that correct?"},
	}
}

func (g *Garcon) addItemToGroupOrder(m slack.Msg) []slack.OutgoingMessage {
	matches, err := findElementsInString(orderPlacingPattern, []string{"item"}, m.Text)
	item := strings.TrimSpace(matches["item"])

	if g.debug {
		log.Printf("I've received an order for '%v', and will be adding it to the group order", item)
	}

	if err != nil || len(item) == 0 {
		return g.genericHelpResponse(m)
	}

	g.Order[g.Patrons[m.User].Name] = item

	t := fmt.Sprintf("Okay @%v, I've got your order.", g.Patrons[m.User].Name)
	return []slack.OutgoingMessage{slack.OutgoingMessage{Channel: m.Channel, Text: t}}
}

func (g *Garcon) orderIsIncorrect(m slack.Msg) []slack.OutgoingMessage {
	t := fmt.Sprintf("Okay, I'll start over.")

	g.Stage = "ordering"
	g.RequestedRestaurant = ""
	g.Order = make(map[string]string)

	return []slack.OutgoingMessage{
		slack.OutgoingMessage{Channel: m.Channel, Text: t},
	}
}

func (g *Garcon) placeOrder(m slack.Msg) []slack.OutgoingMessage {
	restaurantAddress := g.ActualRestaurant.Vicinity
	restaurantPhone := g.ActualRestaurant.FormattedPhoneNumber

	t := "Okay, I'll send this order off!"

	manifest := *ghostmates.NewManifest(g.createOrderString(), "Group Order")
	from := *ghostmates.NewDeliverySpot(g.RequestedRestaurant, restaurantAddress, restaurantPhone)
	to := g.OrderDestination
	quote, err := g.PostmatesClient.GetQuote(from.Address, to.Address)
	if err != nil {
		t = "I'm having some problems placing this order, please check the logs."
		log.Println("Error creating quote")
		log.Fatal(err)
	}

	err = g.PostmatesClient.CreateDelivery(&manifest, &from, to, quote)
	if err != nil {
		t = "I'm having some problems placing this order, please check the logs."
		log.Println("Error creating delivery")
		log.Fatal(err)
	}

	return []slack.OutgoingMessage{slack.OutgoingMessage{Channel: m.Channel, Text: t}}
}

func (g *Garcon) createOrderString() string {
	orders := ""
	for user, order := range g.Order {
		orders = fmt.Sprintf("%v%v: %v\n", orders, strings.Title(user), order)
	}
	return strings.TrimSpace(orders)
}

func (g *Garcon) orderStatusResponse(m slack.Msg) []slack.OutgoingMessage {
	statusTemplate := "Here's what I have for your order from %v:\n```\n%v\n```"
	statusMessage := fmt.Sprintf(statusTemplate, g.RequestedRestaurant, g.createOrderString())
	return []slack.OutgoingMessage{slack.OutgoingMessage{Channel: m.Channel, Text: statusMessage}}
}

func (g *Garcon) genericCancelReponse(m slack.Msg) []slack.OutgoingMessage {
	g.Reset()
	t := "Very well then, I'll disappear for now!"
	return []slack.OutgoingMessage{slack.OutgoingMessage{Channel: m.Channel, Text: t}}
}

// NewGarcon constructs a new instance of Garcon and establishes all the behavior functions
// Note that while it is possible for us to make this code broader and more reusable, I don't
// have an intense interest in doing that right now, and I think such a structure would render
// the code either hilariously unreadable, complicated, and most likely both
func NewGarcon() *Garcon {
	g := &Garcon{
		SelfName: "garcon",
		Stage:    "uninitiated",
		Order:    make(map[string]string),
	}

	g.CommandExamples = map[string][]string{
		"prompted": []string{
			"We'd like to place an order for the Chili's at 45th & Lamar",
			"We would like to order from the Chili's at 45th & Lamar",
		},
		"ordering": []string{
			"@garcon, I'd like a banana",
			"@garcon I'll have the tuna melt",
			"@garcon, what's our order look like so far?",
		},
		"confirmation": []string{
			"yes",
			"no",
		},
		"always": []string{
			"@garcon, go away",
			"@garcon, help!",
		},
	}

	// possible returns: affirmative, negative, additive, cancelling, irrelevant, insufficient
	g.MessageTypeFuncs = map[string]func(slack.Msg) (string, error){
		"uninitiated": func(m slack.Msg) (string, error) {
			if g.cancellationCommandIssued(m) {
				return "cancelling", nil
			}
			if strings.ToLower(m.Text) == "oh, garçon?" || strings.ToLower(m.Text) == "oh, garcon?" {
				return "affirmative", nil
			}
			return "irrelevant", nil
		},
		"prompted": func(m slack.Msg) (string, error) {
			if g.cancellationCommandIssued(m) {
				return "cancelling", nil
			}
			if g.helpRequested(m) {
				return "insufficient", nil
			}
			if responseIsNegative(m.Text) || m.User != g.InterlocutorID {
				return "negative", nil
			}
			if stringFitsPattern(orderInitiationPattern, m.Text) {
				return "affirmative", nil
			}
			return "insufficient", nil
		},
		"ordering": func(m slack.Msg) (string, error) {
			if g.cancellationCommandIssued(m) {
				return "cancelling", nil
			}
			if g.helpRequested(m) {
				return "insufficient", nil
			}
			if g.itemAddedToOrder(m) {
				return "contributing", nil
			}
			if g.orderStatusCheckRequested(m) {
				return "status", nil
			}
			if g.readyToPlaceOrder(m) {
				return "affirmative", nil
			}
			return "indeterminable", nil
		},
		"confirmation": func(m slack.Msg) (string, error) {
			if g.cancellationCommandIssued(m) {
				return "cancelling", nil
			}
			if g.helpRequested(m) {
				return "insufficient", nil
			}
			if responseIsAffirmative(m.Text) {
				return "affirmative", nil
			}
			if responseIsNegative(m.Text) {
				return "negative", nil
			}
			return "insufficient", nil
		},
	}

	g.ReactionFuncs = map[string]map[string]func(slack.Msg) []slack.OutgoingMessage{
		"uninitiated": map[string]func(m slack.Msg) []slack.OutgoingMessage{
			"affirmative": g.helloGarcon,
		},
		"prompted": map[string]func(m slack.Msg) []slack.OutgoingMessage{
			"affirmative":  g.validateRestaurant,
			"insufficient": g.genericHelpResponse,
			"cancelling":   g.genericCancelReponse,
		},
		"ordering": map[string]func(m slack.Msg) []slack.OutgoingMessage{
			"affirmative":  g.validateOrder,
			"contributing": g.addItemToGroupOrder,
			"insufficient": g.genericHelpResponse,
			"cancelling":   g.genericCancelReponse,
			"status":       g.orderStatusResponse,
		},
		"confirmation": map[string]func(m slack.Msg) []slack.OutgoingMessage{
			"affirmative":  g.placeOrder,
			"negative":     g.orderIsIncorrect,
			"cancelling":   g.genericCancelReponse,
			"insufficient": g.genericHelpResponse,
		},
	}

	return g
}
