package orderbook

import (
	"time"
	"testing"

	"github.com/iov-one/weave"
	"github.com/iov-one/weave/coin"
	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/orm"
	"github.com/iov-one/weave/weavetest"
)

func TestMarketIDindexer(t *testing.T) {
	cases := map[string]struct {
		obj orm.Object
		wantErr *errors.Error 
	}{
		"success": {
			obj: orm.NewSimpleObj(nil, new(OrderBook)), 
			wantErr: nil,
		},
		"not orderbook": {
			obj: orm.NewSimpleObj(nil, new(Order)),
			wantErr: errors.ErrState,
		},
	}

	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			if _, err := marketIDindexer(tc.obj); !tc.wantErr.Is(err) {
				t.Fatalf("unexpected error: %+v", err)
			}
		})
	}
}

func TestOpenOrderIndexer(t *testing.T) {
	cases := map[string]struct {
		obj orm.Object
		wantErr *errors.Error 
	}{
		"success": {
			obj: orm.NewSimpleObj(nil, new(Order)), 
			wantErr: nil,
		},
		//TODO BuildOpenOrderIndex cases 
		"failure not order": {
			obj: orm.NewSimpleObj(nil, new(OrderBook)),
			wantErr: errors.ErrState,
		},
	}

	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			if _, err := openOrderIndexer(tc.obj); !tc.wantErr.Is(err) {
				t.Fatalf("unexpected error: %+v", err)
			}
		})
	}
}

func TestOrderIDindexer(t *testing.T) {
	cases := map[string]struct {
		obj orm.Object
		wantErr *errors.Error 
	}{
		"success": {
			obj: orm.NewSimpleObj(nil, new(Trade)), 
			wantErr: nil,
		},
		"not trade": {
			obj: orm.NewSimpleObj(nil, new(OrderBook)),
			wantErr: errors.ErrState,
		},
	}

	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			if _, err := orderIDIndexer(tc.obj); !tc.wantErr.Is(err) {
				t.Fatalf("unexpected error: %+v", err)
			}
		})
	}
}

func TestOrderBookTimedIndexer(t *testing.T) {
	now := weave.AsUnixTime(time.Now())
	invalidTime:= weave.AsUnixTime(time.Unix(-1, 0))

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
				OrderID:     weavetest.SequenceID(14),
				OrderBookID: weavetest.SequenceID(2),
				Taker:       weavetest.NewCondition().Address(),
				Maker:       weavetest.NewCondition().Address(),
				TakerPaid:   coin.NewCoinp(100, 0, "ETH"),
				MakerPaid:   coin.NewCoinp(7, 234456, "BTC"),
				ExecutedAt:  invalidTime,
			}

	cases := map[string]struct {
		obj orm.Object
		wantErr *errors.Error 
	}{
		"success": {
			obj: orm.NewSimpleObj(nil, validTrade), 
			wantErr: nil,
		},
		"not trade": {
			obj: orm.NewSimpleObj(nil, new(OrderBook)),
			wantErr: errors.ErrState,
		},
		"invalid execution time": {
			obj: orm.NewSimpleObj(nil, invalidTrade),
			wantErr: errors.ErrState,
		},
	}

	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			if _, err := orderBookTimedIndexer(tc.obj); !tc.wantErr.Is(err) {
				t.Fatalf("unexpected error: %+v", err)
			}
		})
	}
}
