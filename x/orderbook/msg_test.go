package orderbook

import (
	"testing"

	"github.com/iov-one/weave"
	coin "github.com/iov-one/weave/coin"
	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/weavetest"
)

func TestValidateCreateOrderBookMsg(t *testing.T) {
	cases := map[string]struct {
		msg     weave.Msg
		wantErr *errors.Error
	}{
		"success": {
			msg: &CreateOrderBookMsg{
				Metadata:  &weave.Metadata{Schema: 1},
				MarketID:  weavetest.SequenceID(5),
				AskTicker: "BAR",
				BidTicker: "FOO",
			},
			wantErr: nil,
		},
		"missing metadata": {
			msg: &CreateOrderBookMsg{
				MarketID:  weavetest.SequenceID(5),
				AskTicker: "BAR",
				BidTicker: "FOO",
			},
			wantErr: errors.ErrMetadata,
		},
		"bad market id": {
			msg: &CreateOrderBookMsg{
				Metadata:  &weave.Metadata{Schema: 1},
				MarketID:  []byte{7, 99, 0},
				AskTicker: "BAR",
				BidTicker: "FOO",
			},
			wantErr: errors.ErrInput,
		},
		"missing market id": {
			msg: &CreateOrderBookMsg{
				Metadata:  &weave.Metadata{Schema: 1},
				AskTicker: "BAR",
				BidTicker: "FOO",
			},
			wantErr: errors.ErrEmpty,
		},
		"invalid ask ticker": {
			msg: &CreateOrderBookMsg{
				Metadata:  &weave.Metadata{Schema: 1},
				MarketID:  weavetest.SequenceID(5),
				AskTicker: "LONGER",
				BidTicker: "ZOO",
			},
			wantErr: errors.ErrCurrency,
		},
		"invalid bid ticker": {
			msg: &CreateOrderBookMsg{
				Metadata:  &weave.Metadata{Schema: 1},
				MarketID:  weavetest.SequenceID(5),
				AskTicker: "FOO",
				BidTicker: "M00N",
			},
			wantErr: errors.ErrCurrency,
		},
		"wrong ticker order": {
			msg: &CreateOrderBookMsg{
				Metadata:  &weave.Metadata{Schema: 1},
				MarketID:  weavetest.SequenceID(5),
				AskTicker: "FOO",
				BidTicker: "BAR",
			},
			wantErr: errors.ErrCurrency,
		},
		"missing ticker": {
			msg: &CreateOrderBookMsg{
				Metadata:  &weave.Metadata{Schema: 1},
				MarketID:  weavetest.SequenceID(5),
				AskTicker: "FOO",
			},
			wantErr: errors.ErrCurrency,
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

func TestValidateCancelOrderMsg(t *testing.T) {
	cases := map[string]struct {
		msg     weave.Msg
		wantErr *errors.Error
	}{
		"success": {
			msg: &CancelOrderMsg{
				Metadata: &weave.Metadata{Schema: 1},
				OrderID:  weavetest.SequenceID(5),
			},
			wantErr: nil,
		},
		"missing metadata": {
			msg: &CancelOrderMsg{
				OrderID: weavetest.SequenceID(5),
			},
			wantErr: errors.ErrMetadata,
		},
		"bad order id": {
			msg: &CancelOrderMsg{
				Metadata: &weave.Metadata{Schema: 1},
				OrderID:  []byte{0, 0, 0, 0, 0, 1},
			},
			wantErr: errors.ErrInput,
		},
		"missing order id": {
			msg: &CancelOrderMsg{
				Metadata: &weave.Metadata{Schema: 1},
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

func TestValidateCreateOrderMsg(t *testing.T) {
	trader := weavetest.NewCondition().Address()

	cases := map[string]struct {
		msg     weave.Msg
		wantErr *errors.Error
	}{
		"success": {
			msg: &CreateOrderMsg{
				Metadata:    &weave.Metadata{Schema: 1},
				Trader:      trader,
				OrderBookID: weavetest.SequenceID(12345),
				Offer:       coin.NewCoinp(100, 12345, "ETH"),
				Price:       NewAmountp(11, 0),
			},
			wantErr: nil,
		},
		"missing metadata": {
			msg: &CreateOrderMsg{
				Trader:      trader,
				OrderBookID: weavetest.SequenceID(12345),
				Offer:       coin.NewCoinp(100, 12345, "ETH"),
				Price:       NewAmountp(11, 0),
			},
			wantErr: errors.ErrMetadata,
		},
		"bad trader": {
			msg: &CreateOrderMsg{
				Metadata:    &weave.Metadata{Schema: 1},
				Trader:      []byte{17, 29, 0, 8},
				OrderBookID: weavetest.SequenceID(12345),
				Offer:       coin.NewCoinp(100, 12345, "ETH"),
				Price:       NewAmountp(11, 0),
			},
			wantErr: errors.ErrInput,
		},
		"bad orderbook id": {
			msg: &CreateOrderMsg{
				Metadata:    &weave.Metadata{Schema: 1},
				Trader:      trader,
				OrderBookID: []byte{32, 0, 0, 1},
				Offer:       coin.NewCoinp(100, 12345, "ETH"),
				Price:       NewAmountp(11, 0),
			},
			wantErr: errors.ErrInput,
		},
		"missing offer": {
			msg: &CreateOrderMsg{
				Metadata:    &weave.Metadata{Schema: 1},
				Trader:      trader,
				OrderBookID: weavetest.SequenceID(12345),
				Price:       NewAmountp(11, 0),
			},
			wantErr: errors.ErrEmpty,
		},
		"invalid offer": {
			msg: &CreateOrderMsg{
				Metadata:    &weave.Metadata{Schema: 1},
				Trader:      trader,
				OrderBookID: weavetest.SequenceID(12345),
				Offer:       coin.NewCoinp(1, 0, "1WIN!"),
				Price:       NewAmountp(11, 0),
			},
			wantErr: errors.ErrCurrency,
		},
		"negative offer": {
			msg: &CreateOrderMsg{
				Metadata:    &weave.Metadata{Schema: 1},
				Trader:      trader,
				OrderBookID: weavetest.SequenceID(12345),
				Offer:       coin.NewCoinp(-5, 0, "ETH"),
				Price:       NewAmountp(11, 0),
			},
			wantErr: errors.ErrInput,
		},
		"negative price": {
			msg: &CreateOrderMsg{
				Metadata:    &weave.Metadata{Schema: 1},
				Trader:      trader,
				OrderBookID: weavetest.SequenceID(12345),
				Offer:       coin.NewCoinp(5, 0, "ETH"),
				Price:       NewAmountp(-1, 0),
			},
			wantErr: errors.ErrInput,
		},
		"invalid price": {
			msg: &CreateOrderMsg{
				Metadata:    &weave.Metadata{Schema: 1},
				Trader:      trader,
				OrderBookID: weavetest.SequenceID(12345),
				Offer:       coin.NewCoinp(5, 0, "ETH"),
				Price:       NewAmountp(14, -20),
			},
			wantErr: errors.ErrState,
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
