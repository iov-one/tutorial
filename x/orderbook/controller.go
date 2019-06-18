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
	var other Side
	var descending bool
	var acceptable func(our, their *Amount) bool

	// We configure which orders to search and which direction
	if order.Side == Side_Ask {
		// we want the highest price for our Ask token
		other = Side_Bid
		descending = true
		acceptable = func(our, their *Amount) bool { return !our.Greater(their) }
	} else {
		// we want the lowest price for their Ask token
		other = Side_Ask
		descending = false
		acceptable = func(our, their *Amount) bool { return !their.Greater(our) }
	}

	// get an iterator over all matches
	prefix, err := BuildOpenOrderIndex(order.OrderBookID, order.OrderState, other, nil)
	if err != nil {
		return errors.Wrap(err, "prepare prefix to scan")
	}
	// figure out price and if we want highest or lowest
	matches, err := c.orderBucket.PrefixScan(db, prefix, descending)
	if err != nil {
		return errors.Wrap(err, "prefix scan")
	}

	// process every match until the offer is closed
	for order.RemainingOffer.IsPositive() && matches.Valid() {
		var match Order
		err := matches.Load(&match)
		if err != nil {
			return errors.Wrap(err, "loading match")
		}

		// if price doesn't match, break out of loop
		if !acceptable(order.Price, match.Price) {
			break
		}

		// otherwise, execute trade
		err = c.executeTrade(db, order, &match)
		if err != nil {
			return errors.Wrap(err, "executing trade")
		}
	}

	return nil
}

// execute trade assumes this was already validated as acceptable.
// it uses the price of the counter offer (which is equal to or better
// that what the new order requested).
// the smaller order will be emptied, and the larger one typically
// left with some remaining balance
func (c controller) executeTrade(db weave.KVStore, order *Order, counter *Order) error {
	// TODO
	return nil
}
