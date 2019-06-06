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
