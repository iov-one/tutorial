package orderbook

import (
	"github.com/iov-one/weave"
	"github.com/iov-one/weave/coin"
	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/x/cash"
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
	mover       cash.CoinMover
}

func (c controller) settleOrder(db weave.KVStore, order *Order, now weave.UnixTime) error {
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
		err = c.executeTrade(db, order, &match, now)
		if err != nil {
			return err
		}
	}

	return nil
}

// execute trade assumes this was already validated as acceptable.
// it uses the price of the counter offer (which is equal to or better
// that what the new order requested).
// the smaller order will be emptied, and the larger one typically
// left with some remaining balance
func (c controller) executeTrade(db weave.KVStore, taker, maker *Order, now weave.UnixTime) error {
	ask, bid := taker, maker
	if taker.Side == Side_Bid {
		bid, ask = taker, maker
	}

	askVal, bidVal, err := amountToSettle(ask, bid, maker.Price)
	if err != nil {
		return errors.Wrapf(err, "executing trade %X with %X", taker.ID, maker.ID)
	}

	takerPaid, makerPaid := askVal, bidVal
	if taker.Side == Side_Bid {
		makerPaid, takerPaid = askVal, bidVal
	}

	// create a trade record
	trade := Trade{
		Metadata:    &weave.Metadata{},
		OrderBookID: ask.OrderBookID,
		TakerID:    taker.ID,
		MakerID:    maker.ID,
		Taker:      taker.Trader,
		Maker:      maker.Trader,
		TakerPaid:  takerPaid,
		MakerPaid:  makerPaid,
		ExecutedAt: now,
	}
	if err := c.tradeBucket.Put(db, &trade); err != nil {
		return errors.Wrap(err, "saving trade")
	}

	if err := c.payout(db, ask, bid.Trader, *askVal, now); err != nil {
		return errors.Wrap(err, "ask payout")
	}
	if err := c.payout(db, bid, ask.Trader, *bidVal, now); err != nil {
		return errors.Wrap(err, "bid payout")
	}

	return nil
}

func (c controller) payout(db weave.KVStore, from *Order, to weave.Address, amount coin.Coin, now weave.UnixTime) error {
	// payout the ask side
	err := c.mover.MoveCoins(db, from.Address(), to, amount)
	if err != nil {
		return errors.Wrap(err, "paying trader")
	}
	rem, err := from.RemainingOffer.Subtract(amount)
	if err != nil {
		return errors.Wrap(err, "deducting ask tokens")
	}
	from.RemainingOffer = &rem
	from.UpdatedAt = now
	if from.RemainingOffer.IsZero() {
		// TODO: remember the hanging chads (rr... 1 fractional)
		from.OrderState = OrderState_Done
	}
	if err := c.orderBucket.Put(db, from); err != nil {
		return errors.Wrap(err, "updating order")
	}
	return nil
}

func amountToSettle(ask *Order, bid *Order, price *Amount) (askVal, bidVal *coin.Coin, err error) {
	// TODO: how to handler remainders and rounding?
	// can we be left with an amountToSmall to ever settle?

	askVal = ask.RemainingOffer
	bidVal, err = price.Multiply(askVal)
	if err != nil {
		return
	}

	// if we don't have enough to cover the ask, then we use the bid amount
	if bid.RemainingOffer.Compare(*bidVal) < 0 {
		bidVal = bid.RemainingOffer
		askVal, err = price.Divide(bidVal)
	}

	return
}
