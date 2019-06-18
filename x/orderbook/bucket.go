package orderbook

import (
	"bytes"
	"encoding/binary"

	"github.com/iov-one/tutorial/morm"
	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/orm"
)

const (
	// Assumed maximum ticker letter size is 5
	tickerByteSize = 5
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

// NewOrderBookBucket initates orderbook with required indexes/
// TODO remove marketIDindexer if proven unnecessary
func NewOrderBookBucket() *OrderBookBucket {
	b := morm.NewModelBucket("orderbook", &OrderBook{},
		morm.WithIndex("market", marketIDindexer, false),
		morm.WithIndex("marketWithTickers", marketIDTickersIndexer, true),
	)
	return &OrderBookBucket{
		ModelBucket: b,
	}
}

// marketIDindexer indexes by market id to easily query all books on one market
// This index is somewhat redundant. In future could be removed if we can provide
// still usable client-side API without this
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

// marketIDTickersIndexer produces in SQL parlance, a compound index
// (MarketID, AskTicker, BidTicker) -> index
func marketIDTickersIndexer(obj orm.Object) ([]byte, error) {
	if obj == nil || obj.Value() == nil {
		return nil, nil
	}
	orderbook, ok := obj.Value().(*OrderBook)

	if !ok {
		return nil, errors.Wrapf(errors.ErrState, "expected orderbook, got %T", obj.Value())
	}

	return BuildMarketIDTickersIndex(orderbook), nil
}

// BuildMarketIDTickersIndex indexByteSize = 8(MarketID) + ask ticker size + bid ticker size
func BuildMarketIDTickersIndex(orderbook *OrderBook) []byte {
	askTickerByte := make([]byte, tickerByteSize)
	copy(askTickerByte, orderbook.AskTicker)

	bidTickerByte := make([]byte, tickerByteSize)
	copy(bidTickerByte, orderbook.BidTicker)

	return bytes.Join([][]byte{orderbook.MarketID, askTickerByte, bidTickerByte}, nil)
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
	return BuildOpenOrderIndex(order.OrderBookID, order.OrderState, order.Side, order.Price)
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
func BuildOpenOrderIndex(orderBookID []byte, state OrderState, side Side, price *Amount) ([]byte, error) {
	// we don't index if state isn't open
	if state != OrderState_Open {
		return nil, nil
	}

	res := make([]byte, 9, 9+16)
	copy(res, orderBookID)
	res[8] = byte(side)

	// if price is nil, just return the prefix for scanning
	if price == nil {
		return res, nil
	}

	lex, err := price.Lexographic()
	if err != nil {
		return nil, errors.Wrap(err, "building order index")
	}
	res = append(res, lex...)
	return res, nil
}

type TradeBucket struct {
	morm.ModelBucket
}

func NewTradeBucket() *TradeBucket {
	b := morm.NewModelBucket("trade", &Trade{},
		morm.WithMultiKeyIndex("order", orderIDMultiIndexer, false),
		morm.WithIndex("orderbook", orderBookTimedIndexer, false),
	)
	return &TradeBucket{
		ModelBucket: b,
	}
}

// orderIDIndexer indexes by both maker and taker order id
// to give us easy lookup of all trades that fulfilled a given order
func orderIDMultiIndexer(obj orm.Object) ([][]byte, error) {
	if obj == nil || obj.Value() == nil {
		return nil, nil
	}
	trade, ok := obj.Value().(*Trade)
	if !ok {
		return nil, errors.Wrapf(errors.ErrState, "expected trade, got %T", obj.Value())
	}
	return [][]byte{trade.MakerID, trade.TakerID}, nil
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

// BuildOrderBookTimeIndex produces 8 bytes OrderBookID || big-endian ExecutedAt
// This allows lexographical searches over the time ranges (or earliest or latest)
// of all trades within one orderbook
func BuildOrderBookTimeIndex(trade *Trade) ([]byte, error) {
	res := make([]byte, 16)
	copy(res, trade.OrderBookID)
	// this would violate lexographical ordering as negatives would be highest
	if trade.ExecutedAt < 0 {
		return nil, errors.Wrap(errors.ErrState, "cannot index negative execution times")
	}
	binary.BigEndian.PutUint64(res[8:], uint64(trade.ExecutedAt))
	return res, nil
}
