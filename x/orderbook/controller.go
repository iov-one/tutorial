package orderbook

import (
	"github.com/iov-one/weave"
	"github.com/iov-one/weave/errors"
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

type controller struct {
	orderBucket *OrderBucket
	tradeBucket *TradeBucket
}

func (c controller) settleOrder(db weave.KVStore, order *Order) error {
	other := Side_Ask
	if order.Side == Side_Ask {
		other = Side_Bid
	}

	// get an iterator over all matches
	prefix, err := BuildOpenOrderIndex(order.OrderBookID, order.OrderState, other, nil)
	if err != nil {
		return errors.Wrap(err, "prepare prefix to scan")
	}
	// TODO: figure out price and if we want highest or lowest
	matches, err := c.orderBucket.PrefixScan(db, prefix, false)
	if err != nil {
		return errors.Wrap(err, "prefix scan")
	}

	// process every match
	for order.RemainingOffer.IsPositive() && matches.Valid() {
		var match Order
		err := matches.Load(&match)
		if err != nil {
			return errors.Wrap(err, "loading match")
		}

		// TODO: if price doesn't match, break out of loop

		// TODO: otherwise, settle the trade to the limit of the lower order
	}

	return nil
}
