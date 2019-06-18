package orderbook

import (
	"github.com/iov-one/weave"
)

func orderCondition(id []byte) weave.Condition {
	if len(id) == 0 {
		panic("developer error: must save before taking address")
	}
	return weave.NewCondition("orderbook", "order", id)
}

// Address returns unique address, so an order contract can manage funds
func (order *Order) Address() weave.Address {
	return orderCondition(order.ID).Address()
}
