package main

// Order is our gateway struct between Favor and Garcon,
type Order struct {
	Begun                bool
	RequestedOrders      map[string]string
	RequestedRestauraunt string
	ActualRestaurant     string
}

// NewOrder constructs a new Order instance for us.
func NewOrder() *Order {
	return &Order{}
}
