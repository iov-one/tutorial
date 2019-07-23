package orderbook

import (
	"testing"

	"github.com/iov-one/weave"
	"github.com/iov-one/weave/coin"
	"github.com/iov-one/weave/migration"
	"github.com/iov-one/weave/store"
	"github.com/iov-one/weave/weavetest"
	"github.com/iov-one/weave/weavetest/assert"
	"github.com/iov-one/weave/x/cash"
)

func TestSettleOrder(t *testing.T) {
	db := store.MemStore()

	schema := migration.NewSchemaBucket()
	_, err := schema.Create(db, &migration.Schema{
		Metadata: &weave.Metadata{Schema: 1},
		Pkg:      "cash",
		Version:  1,
	})
	assert.Nil(t, err)

	mover := cash.NewController(cash.NewBucket())
	orders := NewOrderBucket()
	ctrl := controller{
		orderBucket: orders,
		tradeBucket: NewTradeBucket(),
		mover:       mover,
	}
	// now := weave.UnixTime(789)

	maker := weavetest.NewCondition().Address()
	taker := weavetest.NewCondition().Address()

	// make the orderbook
	// Do we need to save this?
	book := &OrderBook{
		ID:        weavetest.SequenceID(5),
		MarketID:  weavetest.SequenceID(2),
		AskTicker: "ASK",
		BidTicker: "BID",
	}

	// set up some existing orders
	makeAsk := weave.UnixTime(500)
	ask := &Order{
		ID:             weavetest.SequenceID(1),
		Trader:         maker,
		OrderBookID:    book.ID,
		IsAsk:          true,
		OrderState:     OrderState_Open,
		OriginalOffer:  coin.NewCoinp(20, 0, "ASK"),
		RemainingOffer: coin.NewCoinp(15, 0, "ASK"),
		Price:          NewAmountp(6, 0), // we want 6 bid for 1 ask (90 bid in total)
		UpdatedAt:      makeAsk,
		CreatedAt:      makeAsk,
	}
	err = orders.Put(db, ask)
	assert.Nil(t, err)
	// issue cash on the contract to cover it
	err = mover.CoinMint(db, ask.Address(), *ask.RemainingOffer)
	assert.Nil(t, err)

	// make a new offer to claim if
	exec := weave.UnixTime(5678)
	bid := &Order{
		ID:             weavetest.SequenceID(1),
		Trader:         taker,
		OrderBookID:    book.ID,
		IsAsk:          false,
		OrderState:     OrderState_Open,
		OriginalOffer:  coin.NewCoinp(66, 4, "BID"), // this should match 11 ask and then 4 fractional returned
		RemainingOffer: coin.NewCoinp(66, 4, "BID"),
		Price:          NewAmountp(6, 0), // we want 6 bid for 1 ask (same price)
		UpdatedAt:      exec,
		CreatedAt:      exec,
	}
	err = orders.Put(db, bid)
	assert.Nil(t, err)
	err = mover.CoinMint(db, bid.Address(), *bid.OriginalOffer)
	assert.Nil(t, err)

	// try to settle it
	err = ctrl.settleOrder(db, bid, exec)
	assert.Nil(t, err)

	// check orders updated

	// check trade created

	// check balances correct

	money, err := mover.Balance(db, ask.Address())
	assert.Nil(t, err)
	assert.Equal(t, coin.Coins{coin.NewCoinp(4, 0, "ASK")}, money)

	money, err = mover.Balance(db, bid.Address())
	assert.Nil(t, err)
	assert.Equal(t, coin.Coins{}, money)

	money, err = mover.Balance(db, maker)
	assert.Nil(t, err)
	assert.Equal(t, coin.Coins{coin.NewCoinp(66, 0, "BID")}, money)

	money, err = mover.Balance(db, taker)
	assert.Nil(t, err)
	assert.Equal(t, coin.Coins{coin.NewCoinp(11, 0, "ASK"), coin.NewCoinp(0, 4, "BID")}, money)

}
