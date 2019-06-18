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

// OrderAddress create an Address controled by an order contract
func OrderAddress(order *Order) weave.Address {
	return orderCondition(order.ID).Address()
}
