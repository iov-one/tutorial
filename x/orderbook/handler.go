package orderbook

import (
	"github.com/iov-one/weave"
	"github.com/iov-one/weave/coin"
	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/migration"
	"github.com/iov-one/weave/x"
	"github.com/iov-one/weave/x/cash"
)

const (
	packageName            = "orderbook"
	newOrderBookCost int64 = 100
	newOrderCost     int64 = 10
	cancelOrderCost  int64 = 10
)

// RegisterQuery registers exchange buckets for querying.
func RegisterQuery(qr weave.QueryRouter) {
	NewMarketBucket().Register("markets", qr)
	NewOrderBookBucket().Register("orderbooks", qr)
	NewOrderBucket().Register("orders", qr)
	NewTradeBucket().Register("trades", qr)
}

// RegisterRoutes registers handlers for orderbook message processing.
func RegisterRoutes(r weave.Registry, auth x.Authenticator, mover cash.CoinMover) {
	r = migration.SchemaMigratingRegistry(packageName, r)

	r.Handle(CreateOrderBookMsg{}.Path(), NewOrderBookHandler(auth))
	r.Handle(CreateOrderMsg{}.Path(), NewCreateOrderHandler(auth, mover))
	r.Handle(CancelOrderMsg{}.Path(), NewCancelOrderHandler(auth, mover))
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

// CreateOrderHandler will handle creating orderbooks
type CreateOrderHandler struct {
	auth            x.Authenticator
	mover           cash.CoinMover
	orderBucket     *OrderBucket
	tradeBucket     *TradeBucket
	orderBookBucket *OrderBookBucket
	marketBucket    *MarketBucket
}

var _ weave.Handler = CreateOrderHandler{}

// NewCreateOrderHandler creates a handler that allows issuer to
// create orderbooks. Only owner/admin of the market can issue
// new orderbooks
func NewCreateOrderHandler(auth x.Authenticator, mover cash.CoinMover) weave.Handler {
	return CreateOrderHandler{
		auth:            auth,
		mover:           mover,
		orderBucket:     NewOrderBucket(),
		tradeBucket:     NewTradeBucket(),
		orderBookBucket: NewOrderBookBucket(),
		marketBucket:    NewMarketBucket(),
	}
}

// Check just verifies it is properly formed and returns
// the cost of executing it.
func (h CreateOrderHandler) Check(ctx weave.Context, db weave.KVStore, tx weave.Tx) (*weave.CheckResult, error) {
	_, _, err := h.validate(ctx, db, tx)
	if err != nil {
		return nil, err
	}

	return &weave.CheckResult{GasAllocated: newOrderCost}, nil
}

// validate does all common pre-processing between Check and Deliver
func (h CreateOrderHandler) validate(ctx weave.Context, db weave.KVStore, tx weave.Tx) (*CreateOrderMsg, *OrderBook, error) {
	var msg CreateOrderMsg

	if err := weave.LoadMsg(tx, &msg); err != nil {
		return nil, nil, errors.Wrap(err, "load msg")
	}

	// Get rules from the orderbook
	var orderBook OrderBook
	if err := h.orderBookBucket.One(db, msg.OrderBookID, &orderBook); err != nil {
		return nil, nil, errors.Wrap(err, "load orderbook")
	}

	// make sure we have a valid ticker
	t := msg.Offer.Ticker
	if t != orderBook.AskTicker && t != orderBook.BidTicker {
		return nil, nil, errors.Wrap(errors.ErrCurrency, "offer ticker not in this orderbook")
	}

	// ensure they can make this trade
	if msg.Trader != nil && !h.auth.HasAddress(ctx, msg.Trader) {
		return nil, nil, errors.Wrap(errors.ErrUnauthorized, "must be authorized by the Trader")
	}

	// TODO: do we want to check the market it belongs to, to check fee payments, etc?

	return &msg, &orderBook, nil
}

// Deliver creates an orderbook and saves if all preconditions are met
func (h CreateOrderHandler) Deliver(ctx weave.Context, db weave.KVStore, tx weave.Tx) (*weave.DeliverResult, error) {
	msg, ob, err := h.validate(ctx, db, tx)
	if err != nil {
		return nil, err
	}

	now, err := weave.BlockTime(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "no block time in header")
	}

	trader := msg.Trader
	if trader == nil {
		trader = x.MainSigner(ctx, h.auth).Address()
	}

	side := Side_Ask
	if msg.Offer.Ticker == ob.BidTicker {
		side = Side_Bid
	}

	order := &Order{
		Metadata:       &weave.Metadata{},
		OrderBookID:    msg.OrderBookID,
		Trader:         trader,
		Side:           side,
		OrderState:     OrderState_Open,
		OriginalOffer:  msg.Offer,
		RemainingOffer: msg.Offer,
		Price:          msg.Price,
		CreatedAt:      weave.AsUnixTime(now),
		UpdatedAt:      weave.AsUnixTime(now),
		TradeIds:       nil,
	}

	// Save Order
	if err := h.orderBucket.Put(db, order); err != nil {
		return nil, err
	}

	// Send money to contract
	// We must calculate the address after saving to have proper auto-generated ID
	err = h.mover.MoveCoins(db, trader, order.Address(), *msg.Offer)
	if err != nil {
		return nil, errors.Wrap(err, "cannot cover order")
	}

	// TODO: run through match making on this, re-save each match

	// we return the new id on creation to enable easier queries
	return &weave.DeliverResult{Data: order.ID}, err
}

// CancelOrderHandler will handle creating orderbooks
type CancelOrderHandler struct {
	auth        x.Authenticator
	mover       cash.CoinMover
	orderBucket *OrderBucket
}

var _ weave.Handler = CreateOrderHandler{}

// NewCancelOrderHandler creates a handler that allows issuer to
// create orderbooks. Only owner/admin of the market can issue
// new orderbooks
func NewCancelOrderHandler(auth x.Authenticator, mover cash.CoinMover) weave.Handler {
	return CancelOrderHandler{
		auth:        auth,
		mover:       mover,
		orderBucket: NewOrderBucket(),
	}
}

// Check just verifies it is properly formed and returns
// the cost of executing it.
func (h CancelOrderHandler) Check(ctx weave.Context, db weave.KVStore, tx weave.Tx) (*weave.CheckResult, error) {
	_, err := h.validate(ctx, db, tx)
	if err != nil {
		return nil, err
	}

	return &weave.CheckResult{GasAllocated: newOrderCost}, nil
}

// validate does all common pre-processing between Check and Deliver
func (h CancelOrderHandler) validate(ctx weave.Context, db weave.KVStore, tx weave.Tx) (*Order, error) {
	var msg CancelOrderMsg

	if err := weave.LoadMsg(tx, &msg); err != nil {
		return nil, errors.Wrap(err, "load msg")
	}

	// Load existing order
	var order Order
	if err := h.orderBucket.One(db, msg.OrderID, &order); err != nil {
		return nil, errors.Wrap(err, "load order")
	}

	// ensure the order is still open
	if order.OrderState != OrderState_Open {
		return nil, errors.Wrap(errors.ErrState, "can only cancel open orders")
	}

	// ensure the trader authorized it
	if !h.auth.HasAddress(ctx, order.Trader) {
		return nil, errors.Wrap(errors.ErrUnauthorized, "must be authorized by the Trader")
	}

	// TODO: do we want to check the market it belongs to, to check fee payments, etc?
	return &order, nil
}

// Deliver creates an orderbook and saves if all preconditions are met
func (h CancelOrderHandler) Deliver(ctx weave.Context, db weave.KVStore, tx weave.Tx) (*weave.DeliverResult, error) {
	order, err := h.validate(ctx, db, tx)
	if err != nil {
		return nil, err
	}

	now, err := weave.BlockTime(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "no block time in header")
	}

	// Return funds to order creator
	err = h.mover.MoveCoins(db, order.Address(), order.Trader, *order.RemainingOffer)
	if err != nil {
		return nil, errors.Wrap(err, "cannot cover order")
	}

	// Mark the order as closed
	order.OrderState = OrderState_Cancel
	order.RemainingOffer = &coin.Coin{Ticker: order.RemainingOffer.Ticker}
	order.UpdatedAt = weave.AsUnixTime(now)
	if err := h.orderBucket.Put(db, order); err != nil {
		return nil, err
	}

	return &weave.DeliverResult{}, err
}
