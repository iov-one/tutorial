package orderbook

import (
	"testing"

	"github.com/iov-one/weave"
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
