package orderbook

import (
	"testing"

	"github.com/iov-one/weave"
	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/migration"
	"github.com/iov-one/weave/store"
	"github.com/iov-one/weave/weavetest"
	"github.com/iov-one/weave/weavetest/assert"
)

type checkErr func(error) bool

func noErr(err error) bool { return err == nil }

func TestCreateOrderbook(t *testing.T) {
	perm := weave.NewCondition("sig", "ed25519", []byte{1, 2, 3})
	perm2 := weave.NewCondition("sig", "ed25519", []byte{4, 5, 6})

	meta := &weave.Metadata{Schema: 1}

	market := &Market{
		// TODO: handle this in the bucket?
		Metadata: meta,
		ID:       weavetest.SequenceID(1),
		Name:     "Main",
		Owner:    perm.Address(),
	}
	market2 := &Market{
		Metadata: meta,
		ID:       weavetest.SequenceID(2),
		Name:     "Copycat",
		Owner:    perm2.Address(),
	}

	cases := map[string]struct {
		signers        []weave.Condition
		initOrderbooks []OrderBook
		msg            weave.Msg
		expected       *OrderBook
		wantCheckErr   *errors.Error
		wantDeliverErr *errors.Error
	}{
		"nil message": {
			wantCheckErr:   errors.ErrState,
			wantDeliverErr: errors.ErrState,
		},
		"empty message": {
			msg:            &CreateOrderBookMsg{},
			wantCheckErr:   errors.ErrMetadata,
			wantDeliverErr: errors.ErrMetadata,
		},
		"unauthorized": {
			signers: []weave.Condition{perm2},
			msg: &CreateOrderBookMsg{
				Metadata:  meta,
				MarketID:  market.ID,
				AskTicker: "BTC",
				BidTicker: "ETH",
			},
			wantCheckErr:   errors.ErrUnauthorized,
			wantDeliverErr: errors.ErrUnauthorized,
		},
		"success": {
			signers: []weave.Condition{perm},
			msg: &CreateOrderBookMsg{
				Metadata:  meta,
				MarketID:  market.ID,
				AskTicker: "BTC",
				BidTicker: "ETH",
			},
			expected: &OrderBook{
				Metadata:  meta,
				ID:        weavetest.SequenceID(1),
				MarketID:  market.ID,
				AskTicker: "BTC",
				BidTicker: "ETH",
			},
		},
		"invalid request (wrong order of tickers)": {
			signers: []weave.Condition{perm},
			msg: &CreateOrderBookMsg{
				Metadata:  meta,
				MarketID:  market.ID,
				AskTicker: "FOO",
				BidTicker: "BAR",
			},
			wantCheckErr:   errors.ErrCurrency,
			wantDeliverErr: errors.ErrCurrency,
		},
	}

	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			auth := &weavetest.Auth{Signers: tc.signers}
			h := NewOrderBookHandler(auth)

			kv := store.MemStore()
			migration.MustInitPkg(kv, packageName)

			bucket := NewMarketBucket()
			err := bucket.Put(kv, market)
			assert.Nil(t, err)
			err = bucket.Put(kv, market2)
			assert.Nil(t, err)

			// TODO: initialize orderbooks

			tx := &weavetest.Tx{Msg: tc.msg}

			if _, err := h.Check(nil, kv, tx); !tc.wantCheckErr.Is(err) {
				t.Fatalf("unexpected check error: %+v", err)
			}
			if _, err := h.Deliver(nil, kv, tx); !tc.wantDeliverErr.Is(err) {
				t.Fatalf("unexpected deliver error: %+v", err)
			}
		})
	}
}
