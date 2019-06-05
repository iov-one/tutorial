package orderbook

import (
	"time"
	"testing"

	"github.com/iov-one/weave"
	"github.com/iov-one/weave/coin"
	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/weavetest"

	"github.com/iov-one/tutorial/morm"
)

func TestValidateOrderBook(t *testing.T) {
	cases := map[string]struct {
		msg     morm.Model
		wantErr *errors.Error
	}{
		"success, no id": {
			msg: &OrderBook{
				Metadata:  &weave.Metadata{Schema: 1},
				MarketID:  weavetest.SequenceID(5),
				AskTicker: "BAR",
				BidTicker: "FOO",
			},
			wantErr: nil,
		},
		"success, with id": {
			msg: &OrderBook{
				Metadata:  &weave.Metadata{Schema: 1},
				ID: weavetest.SequenceID(13),
				MarketID:  weavetest.SequenceID(1),
				AskTicker: "BAR",
				BidTicker: "FOO",
			},
			wantErr: nil,
		},
		"failure no market id": {
			msg: &OrderBook{
				Metadata:  &weave.Metadata{Schema: 1},
				ID: weavetest.SequenceID(13),
				AskTicker: "BAR",
				BidTicker: "FOO",
			},
			wantErr: errors.ErrEmpty,
		},
	}
	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			if err := tc.msg.Validate(); !tc.wantErr.Is(err) {
				t.Fatalf("unexpected error: %+v", err)
			}
		})
	}
}

func TestValidateMarket(t *testing.T) {
	cases := map[string]struct {
		msg     morm.Model
		wantErr *errors.Error
	}{
		"success, no id": {
			msg: &Market{
				Metadata:  &weave.Metadata{Schema: 1},
				Owner:  weavetest.NewCondition().Address(),
				Name: "Fred",
			},
			wantErr: nil,
		},
		"success, with id": {
			msg: &Market{
				Metadata:  &weave.Metadata{Schema: 1},
				ID: weavetest.SequenceID(2),
				Owner:  weavetest.NewCondition().Address(),
				Name: "Fred",
			},
			wantErr: nil,
		},
		"failure bad owner": {
			msg: &Market{
				Metadata:  &weave.Metadata{Schema: 1},
				ID: weavetest.SequenceID(2),
				Owner:  []byte("foobar"),
				Name: "Fred",
			},
			wantErr: errors.ErrInput,
		},
	}
	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			if err := tc.msg.Validate(); !tc.wantErr.Is(err) {
				t.Fatalf("unexpected error: %+v", err)
			}
		})
	}
}

func TestValidateOrder(t *testing.T) {
	now := weave.AsUnixTime(time.Now())

	cases := map[string]struct {
		msg     morm.Model
		wantErr *errors.Error
	}{
		"success, no id": {
			msg: &Order{
				Metadata:  &weave.Metadata{Schema: 1},
				Trader: weavetest.NewCondition().Address(),
				OrderBookID:  weavetest.SequenceID(5),
				Side: Side_Ask,
				OrderState: OrderState_Open,
				OriginalOffer: coin.NewCoinp(100, 0, "ETH"),
				RemainingOffer: coin.NewCoinp(50, 17, "ETH"),
				Price: NewAmountp(121, 0),
				CreatedAt: now,
				UpdatedAt: now,
			},
			wantErr: nil,
		},
		"success, with id": {
			msg: &Order{
				Metadata:  &weave.Metadata{Schema: 1},
				ID: weavetest.SequenceID(17),
				Trader: weavetest.NewCondition().Address(),
				OrderBookID:  weavetest.SequenceID(5),
				Side: Side_Ask,
				OrderState: OrderState_Open,
				OriginalOffer: coin.NewCoinp(100, 0, "ETH"),
				RemainingOffer: coin.NewCoinp(50, 17, "ETH"),
				Price: NewAmountp(121, 0),
				CreatedAt: now,
				UpdatedAt: now,
			},
			wantErr: nil,
		},
		"missing timestamps": {
			msg: &Order{
				Metadata:  &weave.Metadata{Schema: 1},
				ID: weavetest.SequenceID(17),
				Trader: weavetest.NewCondition().Address(),
				OrderBookID:  weavetest.SequenceID(5),
				Side: Side_Ask,
				OrderState: OrderState_Open,
				OriginalOffer: coin.NewCoinp(100, 0, "ETH"),
				RemainingOffer: coin.NewCoinp(50, 17, "ETH"),
				Price: NewAmountp(121, 0),
			},
			wantErr: errors.ErrEmpty,
		},
	}
	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			if err := tc.msg.Validate(); !tc.wantErr.Is(err) {
				t.Fatalf("unexpected error: %+v", err)
			}
		})
	}
}

func TestValidateTrade(t *testing.T) {
	now := weave.AsUnixTime(time.Now())

	cases := map[string]struct {
		msg     morm.Model
		wantErr *errors.Error
	}{
		"success, no id": {
			msg: &Trade{
				Metadata:  &weave.Metadata{Schema: 1},
				OrderID:  weavetest.SequenceID(14),
				OrderBookID:  weavetest.SequenceID(2),
				Taker: weavetest.NewCondition().Address(),
				Maker: weavetest.NewCondition().Address(),
				TakerPaid: coin.NewCoinp(100, 0, "ETH"),
				MakerPaid: coin.NewCoinp(7, 234456, "BTC"),
				ExecutedAt: now,
			},
			wantErr: nil,
		},
		"success, with id": {
			msg: &Trade{
				Metadata:  &weave.Metadata{Schema: 1},
				ID:  weavetest.SequenceID(7654),
				OrderID:  weavetest.SequenceID(14),
				OrderBookID:  weavetest.SequenceID(2),
				Taker: weavetest.NewCondition().Address(),
				Maker: weavetest.NewCondition().Address(),
				TakerPaid: coin.NewCoinp(100, 0, "ETH"),
				MakerPaid: coin.NewCoinp(7, 234456, "BTC"),
				ExecutedAt: now,
			},
			wantErr: nil,
		},
		"missing payment": {
			msg: &Trade{
				Metadata:  &weave.Metadata{Schema: 1},
				ID:  weavetest.SequenceID(7654),
				OrderID:  weavetest.SequenceID(14),
				OrderBookID:  weavetest.SequenceID(2),
				Taker: weavetest.NewCondition().Address(),
				Maker: weavetest.NewCondition().Address(),
				TakerPaid: coin.NewCoinp(100, 0, "ETH"),
				ExecutedAt: now,
			},
			wantErr: errors.ErrEmpty,
		},
	}
	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			if err := tc.msg.Validate(); !tc.wantErr.Is(err) {
				t.Fatalf("unexpected error: %+v", err)
			}
		})
	}
}
