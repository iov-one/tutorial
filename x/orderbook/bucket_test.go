package orderbook

import (
	"testing"
	"time"

	"github.com/iov-one/weave"
	"github.com/iov-one/weave/coin"
	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/orm"
	"github.com/iov-one/weave/weavetest"
	"github.com/iov-one/weave/weavetest/assert"
)

func TestMarketIDindexer(t *testing.T) {
	marketID := weavetest.SequenceID(5)

	orderbook := &OrderBook{
		Metadata:  &weave.Metadata{Schema: 1},
		MarketID:  marketID,
		AskTicker: "BAR",
		BidTicker: "FOO",
	}

	cases := map[string]struct {
		obj      orm.Object
		expected []byte
		wantErr  *errors.Error
	}{
		"success": {
			obj:      orm.NewSimpleObj(nil, orderbook),
			expected: marketID,
			wantErr:  nil,
		},
		"failure, obj is nil": {
			obj:      nil,
			expected: nil,
			wantErr:  nil,
		},
		// TODO add obj.value nil case
		"not orderbook": {
			obj:      orm.NewSimpleObj(nil, new(Order)),
			expected: nil,
			wantErr:  errors.ErrState,
		},
	}

	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			index, err := marketIDindexer(tc.obj)

			if !tc.wantErr.Is(err) {
				t.Fatalf("unexpected error: %+v", err)
			}
			assert.Equal(t, tc.expected, index)
		})
	}
}

func TestMarketIDTickersIndexer(t *testing.T) {
	marketID := weavetest.SequenceID(5)

	// Orderbook with 3 letter tickers
	orderbook1 := &OrderBook{
		Metadata:  &weave.Metadata{Schema: 1},
		MarketID:  marketID,
		AskTicker: "BAR",
		BidTicker: "FOO",
	}

	expectedIndex1 := []byte{0, 0, 0, 0, 0, 0, 0, 5, 66, 65, 82, 0, 0, 70, 79, 79, 0, 0}

	// Orderbook with 4 letter tickers
	orderbook2 := &OrderBook{
		Metadata:  &weave.Metadata{Schema: 1},
		MarketID:  marketID,
		AskTicker: "BARZ",
		BidTicker: "FOOL",
	}

	expectedIndex2 := []byte{0, 0, 0, 0, 0, 0, 0, 5, 66, 65, 82, 90, 0, 70, 79, 79, 76, 0}

	cases := map[string]struct {
		obj      orm.Object
		expected []byte
		wantErr  *errors.Error
	}{
		"success, with 3 letter tickers": {
			obj:      orm.NewSimpleObj(nil, orderbook1),
			expected: expectedIndex1,
			wantErr:  nil,
		},
		"success, with 4 letter tickers": {
			obj:      orm.NewSimpleObj(nil, orderbook2),
			expected: expectedIndex2,
			wantErr:  nil,
		},
		"failure, obj is nil": {
			obj:      nil,
			expected: nil,
			wantErr:  nil,
		},
		// TODO add obj.value nil case
		"not orderbook": {
			obj:      orm.NewSimpleObj(nil, new(Order)),
			expected: nil,
			wantErr:  errors.ErrState,
		},
	}

	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			index, err := marketIDTickersIndexer(tc.obj)

			if !tc.wantErr.Is(err) {
				t.Fatalf("unexpected error: %+v", err)
			}
			assert.Equal(t, tc.expected, index)
		})
	}
}

func TestOpenOrderIndexer(t *testing.T) {
	now := weave.AsUnixTime(time.Now())
	onceUponATime := weave.AsUnixTime(time.Time{})

	orderBookID := weavetest.SequenceID(5)
	side := Side_Ask

	openOrder := &Order{
		Metadata:       &weave.Metadata{Schema: 1},
		Trader:         weavetest.NewCondition().Address(),
		OrderBookID:    orderBookID,
		Side:           side,
		OrderState:     OrderState_Open,
		OriginalOffer:  coin.NewCoinp(100, 0, "ETH"),
		RemainingOffer: coin.NewCoinp(50, 17, "ETH"),
		Price:          NewAmountp(121, 2125),
		CreatedAt:      onceUponATime,
		UpdatedAt:      now,
	}

	doneOrder := &Order{
		Metadata:       &weave.Metadata{Schema: 1},
		Trader:         weavetest.NewCondition().Address(),
		OrderBookID:    orderBookID,
		Side:           side,
		OrderState:     OrderState_Done,
		OriginalOffer:  coin.NewCoinp(100, 0, "ETH"),
		RemainingOffer: coin.NewCoinp(50, 17, "ETH"),
		Price:          NewAmountp(121, 2125),
		CreatedAt:      onceUponATime,
		UpdatedAt:      now,
	}

	cancelledOrder := &Order{
		Metadata:       &weave.Metadata{Schema: 1},
		Trader:         weavetest.NewCondition().Address(),
		OrderBookID:    orderBookID,
		Side:           side,
		OrderState:     OrderState_Cancel,
		OriginalOffer:  coin.NewCoinp(100, 0, "ETH"),
		RemainingOffer: coin.NewCoinp(50, 17, "ETH"),
		Price:          NewAmountp(121, 2125),
		CreatedAt:      onceUponATime,
		UpdatedAt:      now,
	}

	successCaseExpectedValue := []byte{0, 0, 0, 0, 0, 0, 0, 5, 1, 0, 0, 0, 0, 0, 0, 0, 121, 0, 0, 0, 0, 0, 0, 8, 77}

	cases := map[string]struct {
		obj      orm.Object
		expected []byte
		wantErr  *errors.Error
	}{
		"success": {
			obj:      orm.NewSimpleObj(nil, openOrder),
			expected: successCaseExpectedValue,
			wantErr:  nil,
		},
		"failure, order state done": {
			obj:      orm.NewSimpleObj(nil, doneOrder),
			expected: nil,
			wantErr:  nil,
		},
		"failure, order state cancel": {
			obj:      orm.NewSimpleObj(nil, cancelledOrder),
			expected: nil,
			wantErr:  nil,
		},
		"failure, obj is nil": {
			obj:      nil,
			expected: nil,
			wantErr:  nil,
		},
		// TODO add obj.Value() is nil case
		"failure not order": {
			obj:      orm.NewSimpleObj(nil, new(OrderBook)),
			expected: nil,
			wantErr:  errors.ErrState,
		},
	}

	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			index, err := openOrderIndexer(tc.obj)

			if !tc.wantErr.Is(err) {
				t.Fatalf("unexpected error: %+v", err)
			}
			assert.Equal(t, tc.expected, index)
		})
	}
}

func TestOrderIDindexer(t *testing.T) {
	now := weave.AsUnixTime(time.Now())

	trade := &Trade{
		Metadata:    &weave.Metadata{Schema: 1},
		OrderID:     weavetest.SequenceID(14),
		OrderBookID: weavetest.SequenceID(2),
		Taker:       weavetest.NewCondition().Address(),
		Maker:       weavetest.NewCondition().Address(),
		TakerPaid:   coin.NewCoinp(100, 0, "ETH"),
		MakerPaid:   coin.NewCoinp(7, 234456, "BTC"),
		ExecutedAt:  now,
	}

	cases := map[string]struct {
		obj      orm.Object
		expected []byte
		wantErr  *errors.Error
	}{
		"success": {
			obj:      orm.NewSimpleObj(nil, trade),
			expected: trade.OrderID,
			wantErr:  nil,
		},
		"failure, obj is nil": {
			obj:      nil,
			expected: nil,
			wantErr:  nil,
		},
		// TODO add obj.Value() nil case
		"not trade": {
			obj:      orm.NewSimpleObj(nil, new(OrderBook)),
			expected: nil,
			wantErr:  errors.ErrState,
		},
	}

	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			index, err := orderIDIndexer(tc.obj)

			if !tc.wantErr.Is(err) {
				t.Fatalf("unexpected error: %+v", err)
			}
			assert.Equal(t, tc.expected, index)
		})
	}
}

func TestOrderBookTimedIndexer(t *testing.T) {
	now := weave.AsUnixTime(time.Unix(1, 0))
	invalidTime := weave.AsUnixTime(time.Unix(-1, 0))

	validTrade := &Trade{
		Metadata:    &weave.Metadata{Schema: 1},
		OrderID:     weavetest.SequenceID(14),
		OrderBookID: weavetest.SequenceID(2),
		Taker:       weavetest.NewCondition().Address(),
		Maker:       weavetest.NewCondition().Address(),
		TakerPaid:   coin.NewCoinp(100, 0, "ETH"),
		MakerPaid:   coin.NewCoinp(7, 234456, "BTC"),
		ExecutedAt:  now,
	}

	invalidTrade := &Trade{
		Metadata:    &weave.Metadata{Schema: 1},
		OrderID:     weavetest.SequenceID(13),
		OrderBookID: weavetest.SequenceID(2),
		Taker:       weavetest.NewCondition().Address(),
		Maker:       weavetest.NewCondition().Address(),
		TakerPaid:   coin.NewCoinp(100, 0, "ETH"),
		MakerPaid:   coin.NewCoinp(7, 234456, "BTC"),
		ExecutedAt:  invalidTime,
	}

	successCaseExpectedValue := []byte{0, 0, 0, 0, 0, 0, 0, 14, 0, 0, 0, 0, 0, 0, 0, 1}

	cases := map[string]struct {
		obj      orm.Object
		expected []byte
		wantErr  *errors.Error
	}{
		"success": {
			obj:      orm.NewSimpleObj(nil, validTrade),
			expected: successCaseExpectedValue,
			wantErr:  nil,
		},
		"not a trade": {
			obj:      orm.NewSimpleObj(nil, new(OrderBook)),
			expected: nil,
			wantErr:  errors.ErrState,
		},
		"failure, obj is nil": {
			obj:      nil,
			expected: nil,
			wantErr:  nil,
		},
		// TODO add obj.Value() nil case
		"invalid execution time": {
			obj:      orm.NewSimpleObj(nil, invalidTrade),
			expected: nil,
			wantErr:  errors.ErrState,
		},
	}

	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			index, err := orderBookTimedIndexer(tc.obj)

			if !tc.wantErr.Is(err) {
				t.Fatalf("unexpected error: %+v", err)
			}

			assert.Equal(t, tc.expected, index)
		})
	}
}
