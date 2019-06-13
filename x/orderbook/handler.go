package orderbook

import (
	"github.com/iov-one/weave"
	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/migration"
	"github.com/iov-one/weave/x"
)

const (
	packageName            = "orderbook"
	newOrderBookCost int64 = 100
)

// RegisterQuery registers exchange buckets for querying.
func RegisterQuery(qr weave.QueryRouter) {
	NewMarketBucket().Register("markets", qr)
	NewOrderBookBucket().Register("orderbooks", qr)
}

// RegisterRoutes registers handlers for orderbook message processing.
func RegisterRoutes(r weave.Registry, auth x.Authenticator) {
	r = migration.SchemaMigratingRegistry(packageName, r)

	r.Handle(CreateOrderBookMsg{}.Path(), NewOrderBookHandler(auth))
}

// ------------------- ORDERBOOK HANDLER -------------------

// OrderBookHandler will handle creating orderbooks
type OrderBookHandler struct {
	auth            x.Authenticator
	orderBookBucket *OrderBookBucket
	marketBucket    *MarketBucket
}

var _ weave.Handler = OrderBookHandler{}

// NewOrderBookHandler creates a handler that allows issuer to
// create orderbooks. Only owner/admin of the market can issue
// new orderbooks
func NewOrderBookHandler(auth x.Authenticator) weave.Handler {
	return OrderBookHandler{
		auth:            auth,
		orderBookBucket: NewOrderBookBucket(),
		marketBucket:    NewMarketBucket(),
	}
}

// Check just verifies it is properly formed and returns
// the cost of executing it.
func (h OrderBookHandler) Check(ctx weave.Context, db weave.KVStore, tx weave.Tx) (*weave.CheckResult, error) {
	_, err := h.validate(ctx, db, tx)
	if err != nil {
		return nil, err
	}

	return &weave.CheckResult{GasAllocated: newOrderBookCost}, nil
}

// validate does all common pre-processing between Check and Deliver
func (h OrderBookHandler) validate(ctx weave.Context, db weave.KVStore, tx weave.Tx) (*CreateOrderBookMsg, error) {
	var msg CreateOrderBookMsg

	if err := weave.LoadMsg(tx, &msg); err != nil {
		return nil, errors.Wrap(err, "load msg")
	}

	// Check market with msg.MarketID exists
	var market Market
	if err := h.marketBucket.One(db, msg.MarketID, &market); err != nil {
		return nil, err
	}

	// And ensure the owner has authorized this change
	if !h.auth.HasAddress(ctx, market.Owner) {
		return nil, errors.Wrap(errors.ErrUnauthorized, "only market owner can create orderbook")
	}

	return &msg, nil
}

// Deliver creates an orderbook and saves if all preconditions are met
func (h OrderBookHandler) Deliver(ctx weave.Context, db weave.KVStore, tx weave.Tx) (*weave.DeliverResult, error) {
	msg, err := h.validate(ctx, db, tx)
	if err != nil {
		return nil, err
	}

	//make the orderbook
	orderbook := &OrderBook{
		Metadata:      &weave.Metadata{},
		MarketID:      msg.MarketID,
		AskTicker:     msg.AskTicker,
		BidTicker:     msg.BidTicker,
		TotalAskCount: 0,
		TotalBidCount: 0,
	}

	// the unique index "marketWithTickers" ensures there are no duplicates, would return error here
	if err := h.orderBookBucket.Put(db, orderbook); err != nil {
		return nil, err
	}

	// we return the new id on creation to enable easier queries
	return &weave.DeliverResult{Data: orderbook.ID}, err
}
