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
		ID:    weavetest.SequenceID(1),
		Name:  "Main",
		Owner: perm.Address(),
	}
	market2 := &Market{
		ID:    weavetest.SequenceID(2),
		Name:  "Copycat",
		Owner: perm2.Address(),
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
				Metadata:  &weave.Metadata{},
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
		"matching orderbook already exists": {
			signers: []weave.Condition{perm},
			initOrderbooks: []OrderBook{{
				MarketID:  market.ID,
				AskTicker: "BAR",
				BidTicker: "FOO",
			}},
			msg: &CreateOrderBookMsg{
				Metadata:  meta,
				MarketID:  market.ID,
				AskTicker: "BAR",
				BidTicker: "FOO",
			},
			wantDeliverErr: errors.ErrDuplicate,
		},
		"matching orderbook already exists in other market": {
			signers: []weave.Condition{perm},
			initOrderbooks: []OrderBook{{
				MarketID:  market2.ID,
				AskTicker: "BAR",
				BidTicker: "FOO",
			}},
			msg: &CreateOrderBookMsg{
				Metadata:  meta,
				MarketID:  market.ID,
				AskTicker: "BAR",
				BidTicker: "FOO",
			},
			expected: &OrderBook{
				Metadata:  &weave.Metadata{},
				ID:        weavetest.SequenceID(2),
				MarketID:  market.ID,
				AskTicker: "BAR",
				BidTicker: "FOO",
			},
		},
	}

	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			auth := &weavetest.Auth{Signers: tc.signers}
			h := NewOrderBookHandler(auth)

			kv := store.MemStore()
			migration.MustInitPkg(kv, packageName)

			// initialize markets and orderbooks for test state
			bucket := NewMarketBucket()
			err := bucket.Put(kv, market)
			assert.Nil(t, err)
			err = bucket.Put(kv, market2)
			assert.Nil(t, err)

			orders := NewOrderBookBucket()
			for _, ob := range tc.initOrderbooks {
				err := orders.Put(kv, &ob)
				assert.Nil(t, err)
			}

			tx := &weavetest.Tx{Msg: tc.msg}

			if _, err := h.Check(nil, kv, tx); !tc.wantCheckErr.Is(err) {
				t.Logf("want: %+v", tc.wantCheckErr)
				t.Logf("got: %+v", err)
				t.Fatalf("check (%T)", tc.msg)
			}
			dres, err := h.Deliver(nil, kv, tx)
			if !tc.wantDeliverErr.Is(err) {
				t.Logf("want: %+v", tc.wantDeliverErr)
				t.Logf("got: %+v", err)
				t.Fatalf("check (%T)", tc.msg)
			}

			// TODO: check expected
			if tc.expected != nil {
				var stored OrderBook
				err = orders.One(kv, dres.Data, &stored)
				assert.Nil(t, err)
				assert.Equal(t, tc.expected, &stored)
			}
		})
	}
}
