package orderbook

import (
	"github.com/iov-one/tutorial/morm"
	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/orm"
)

type MarketBucket struct {
	morm.ModelBucket
}

func NewMarketBucket() *MarketBucket {
	b := morm.NewModelBucket("market", &Market{})
	return &MarketBucket{
		ModelBucket: b,
	}
}

type OrderBookBucket struct {
	morm.ModelBucket
}

func NewOrderBookBucket() *OrderBookBucket {
	b := morm.NewModelBucket("orderbook", &OrderBook{},
		morm.WithIndex("market", marketIDindexer, false),
	)
	return &OrderBookBucket{
		ModelBucket: b,
	}
}

// this indexes by market id to easily query all books on one market
func marketIDindexer(obj orm.Object) ([]byte, error) {
	if obj == nil || obj.Value() == nil {
		return nil, nil
	}
	ob, ok := obj.Value().(*OrderBook)
	if !ok {
		return nil, errors.Wrapf(errors.ErrState, "expected orderbook, got %T", obj.Value())
	}
	return ob.MarketID, nil
}

type OrderBucket struct {
	morm.ModelBucket
}

func NewOrderBucket() *OrderBucket {
	b := morm.NewModelBucket("order", &Order{},
		morm.WithIndex("open", openOrderIDindexer, false),
	)
	return &OrderBucket{
		ModelBucket: b,
	}
}

// in SQL parlance, this is a compound index:
//
//   (OrderBookID, Side, Price) WHERE order.OrderState = Open
//
// The purpose is to enable range proofs over price for matching order...
// eg. (orderbook=7, side=ask) and then and Iterate over prices Ascending
// (TODO: add a proper Iterator/First method to ModelBucket - key and index)
func openOrderIDindexer(obj orm.Object) ([]byte, error) {
	if obj == nil || obj.Value() == nil {
		return nil, nil
	}
	order, ok := obj.Value().(*Order)
	if !ok {
		return nil, errors.Wrapf(errors.ErrState, "expected order, got %T", obj.Value())
	}
	return BuildOpenOrderIndex(order)
}

// Goal is compound index like:
//
//   (OrderBookID, Side, Price) WHERE order.OrderState = Open
//
// Stored as - 8 bytes bigendian OrderBookID, 1 byte Side, 8 byte bigendian Price.Whole, 8 byte bigendian Price.Fractional
// We use Price.Lexographic() to produce a lexographic ordering, such than
//
//   A.Lexographic() < B.Lexographic == A < B
//
// This is a very nice trick to get clean range queries over sensible value ranges in a key-value store
func BuildOpenOrderIndex(order *Order) ([]byte, error) {
	// we don't index if state isn't open
	if order.OrderState != OrderState_Open {
		return nil, nil
	}

	res := make([]byte, 9, 9+16)
	copy(res, order.OrderBookID)
	res[8] = byte(order.Side)
	lex, err := order.Price.Lexographic()
	if err != nil {
		return nil, errors.Wrap(err, "building order index")
	}
	copy(res[9:], lex)
	return res, nil
}
