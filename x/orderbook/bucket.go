package orderbook

import (
	"encoding/binary"

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

// marketIDindexer indexes by market id to easily query all books on one market
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
		morm.WithIndex("open", openOrderIndexer, false),
	)
	return &OrderBucket{
		ModelBucket: b,
	}
}

// openOrderIndexer produces, in SQL parlance, a compound index:
//
//   (OrderBookID, Side, Price) WHERE order.OrderState = Open
//
// The purpose is to enable range proofs over price for matching order...
// eg. (orderbook=7, side=ask) and then and Iterate over prices Ascending
// (TODO: add a proper Iterator/First method to ModelBucket - key and index)
func openOrderIndexer(obj orm.Object) ([]byte, error) {
	if obj == nil || obj.Value() == nil {
		return nil, nil
	}
	order, ok := obj.Value().(*Order)
	if !ok {
		return nil, errors.Wrapf(errors.ErrState, "expected order, got %T", obj.Value())
	}
	return BuildOpenOrderIndex(order)
}

// BuildOpenOrderIndex produces a compound index like:
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

	res := make([]byte, 9+16)
	copy(res, order.OrderBookID)
	res[8] = byte(order.Side)
	lex, err := order.Price.Lexographic()
	if err != nil {
		return nil, errors.Wrap(err, "building order index")
	}
	copy(res[9:], lex)
	return res, nil
}

type TradeBucket struct {
	morm.ModelBucket
}

func NewTradeBucket() *TradeBucket {
	b := morm.NewModelBucket("trade", &Trade{},
		morm.WithIndex("order", orderIDIndexer, false),
		morm.WithIndex("orderbook", orderBookTimedIndexer, false),
	)
	return &TradeBucket{
		ModelBucket: b,
	}
}

// orderIDIndexer indexes by order id to give us easy lookup of all trades
// that fulfilled a given order
func orderIDIndexer(obj orm.Object) ([]byte, error) {
	if obj == nil || obj.Value() == nil {
		return nil, nil
	}
	trade, ok := obj.Value().(*Trade)
	if !ok {
		return nil, errors.Wrapf(errors.ErrState, "expected trade, got %T", obj.Value())
	}
	return trade.OrderID, nil
}

// orderBookTimedIndexer indexes trades by
//   (order book id, executed_at)
// so give us easy lookup of the most recently executed trades on a given orderbook
// (we can also use this client side with range queries to select all trades on a given
// book during any given timeframe)
func orderBookTimedIndexer(obj orm.Object) ([]byte, error) {
	if obj == nil || obj.Value() == nil {
		return nil, nil
	}
	trade, ok := obj.Value().(*Trade)
	if !ok {
		return nil, errors.Wrapf(errors.ErrState, "expected trade, got %T", obj.Value())
	}

	return BuildOrderBookTimeIndex(trade)
}

// BuildOpenOrderIndex produces 8 bytes OrderBookID || big-endian ExecutedAt
// This allows lexographical searches over the time ranges (or earliest or latest)
// of all trades within one orderbook
func BuildOrderBookTimeIndex(trade *Trade) ([]byte, error) {
	res := make([]byte, 16)
	copy(res, trade.OrderID)
	// this would violate lexographical ordering as negatives would be highest
	if trade.ExecutedAt < 0 {
		return nil, errors.Wrap(errors.ErrState, "cannot index negative execution times")
	}
	binary.BigEndian.PutUint64(res[8:], uint64(trade.ExecutedAt))
	return res, nil
}
