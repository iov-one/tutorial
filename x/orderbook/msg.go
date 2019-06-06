package orderbook

import (
	"github.com/iov-one/weave"
	coin "github.com/iov-one/weave/coin"
	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/migration"
)

func init() {
	// Migration needs to be registered for every message introduced in the codec.
	// This is the convention to message versioning.
	migration.MustRegister(1, &CreateOrderBookMsg{}, migration.NoModification)
	migration.MustRegister(1, &CreateOrderMsg{}, migration.NoModification)
	migration.MustRegister(1, &CancelOrderMsg{}, migration.NoModification)
}

const (
	pathCreateOrderBook = "order/create_book"
	pathCreateOrder     = "order/create"
	pathCancelOrder     = "order/cancel"
)

var _ weave.Msg = (*CreateOrderBookMsg)(nil)
var _ weave.Msg = (*CreateOrderMsg)(nil)
var _ weave.Msg = (*CancelOrderMsg)(nil)

// ROUTING, Path method fulfills weave.Msg interface to allow routing

func (CreateOrderBookMsg) Path() string {
	return pathCreateOrderBook
}

func (CreateOrderMsg) Path() string {
	return pathCreateOrder
}

func (CancelOrderMsg) Path() string {
	return pathCancelOrder
}

func (m CreateOrderBookMsg) Validate() error {
	if err := m.Metadata.Validate(); err != nil {
		return errors.Wrap(err, "metadata")
	}
	if err := validID(m.MarketID); err != nil {
		return errors.Wrap(err, "market id")
	}
	if !coin.IsCC(m.AskTicker) {
		return errors.Wrapf(errors.ErrCurrency, "Invalid Ask Ticker: %s", m.AskTicker)
	}
	if !coin.IsCC(m.BidTicker) {
		return errors.Wrapf(errors.ErrCurrency, "Invalid Bid Ticker: %s", m.BidTicker)
	}
	if m.BidTicker <= m.AskTicker {
		return errors.Wrapf(errors.ErrCurrency, "ask (%s) must be before bid (%s)", m.AskTicker, m.BidTicker)
	}
	return nil
}

func (m CreateOrderMsg) Validate() error {
	if err := m.Metadata.Validate(); err != nil {
		return errors.Wrap(err, "metadata")
	}
	if err := m.Trader.Validate(); err != nil {
		return errors.Wrap(err, "trader id")
	}
	if err := validID(m.OrderBookID); err != nil {
		return errors.Wrap(err, "orderbook id")
	}
	if m.Offer == nil {
		return errors.Wrap(errors.ErrEmpty, "offer")
	}
	if err := m.Offer.Validate(); err != nil {
		return errors.Wrap(err, "offer")
	}
	if !m.Offer.IsPositive() {
		return errors.Wrap(errors.ErrInput, "offer must be positive")
	}
	if err := m.Price.Validate(); err != nil {
		return errors.Wrap(err, "price")
	}
	if !m.Price.IsPositive() {
		return errors.Wrap(errors.ErrInput, "price must be positive")
	}
	return nil
}

func (m CancelOrderMsg) Validate() error {
	if err := m.Metadata.Validate(); err != nil {
		return errors.Wrap(err, "metadata")
	}
	if err := validID(m.OrderID); err != nil {
		return errors.Wrap(err, "order id")
	}
	return nil
}

// validID returns an error if this is not an 8-byte ID
// as expected for orm.IDGenBucket
func validID(id []byte) error {
	if len(id) == 0 {
		return errors.Wrap(errors.ErrEmpty, "id missing")
	}
	if len(id) != 8 {
		return errors.Wrap(errors.ErrInput, "id is invalid length (expect 8 bytes)")
	}
	return nil
}
