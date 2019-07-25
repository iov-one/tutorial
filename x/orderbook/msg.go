package orderbook

import (
	"fmt"

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

var _ weave.Msg = (*CreateOrderBookMsg)(nil)
var _ weave.Msg = (*CreateOrderMsg)(nil)
var _ weave.Msg = (*CancelOrderMsg)(nil)

// ROUTING, Path method fulfills weave.Msg interface to allow routing

func (CreateOrderBookMsg) Path() string {
	return "order/create_book"
}

func (CreateOrderMsg) Path() string {
	return "order/create"
}

func (CancelOrderMsg) Path() string {
	return "order/cancel"
}

func (m CreateOrderBookMsg) Validate() error {
	var errs error

	errs = errors.AppendField(errs, "metadata", m.Metadata.Validate())
	errs = errors.AppendField(errs, "market id", validID(m.MarketID))

	if !coin.IsCC(m.AskTicker) {
		errs = errors.Append(errs,
			errors.Field("AskTicker", errors.ErrCurrency, fmt.Sprintf("Invalid ask ticker: %s", m.AskTicker)))
	}
	if !coin.IsCC(m.BidTicker) {
		errs = errors.Append(errs,
			errors.Field("BidTicker", errors.ErrCurrency, "Invalid bid ticker"))
	}
	if m.BidTicker <= m.AskTicker {
		errs = errors.Append(errs,
			errors.Field("BidTicker", errors.ErrCurrency, "ask (%s) must be before bid (%s)"))
	}
	return errs
}

func (m CreateOrderMsg) Validate() error {
	var errs error

	errs = errors.AppendField(errs, "metadata", m.Metadata.Validate())
	errs = errors.AppendField(errs, "trader id", m.Trader.Validate())
	errs = errors.AppendField(errs, "orderbook id", validID(m.OrderBookID))

	if m.Offer == nil {
		errs = errors.Append(errs,
			errors.Field("Offer", errors.ErrEmpty, "empty offer"))
	} else if err := m.Offer.Validate(); err != nil {
		errs = errors.AppendField(errs, "Offer", err)
	} else if !m.Offer.IsPositive() {
		errs = errors.Append(errs,
			errors.Field("Offer", errors.ErrInput, "offer must be positive"))
	}

	if err := m.Price.Validate(); err != nil {
		errs = errors.AppendField(errs, "Price", err)
	} else if !m.Price.IsPositive() {
		errs = errors.Append(errs,
			errors.Field("Price", errors.ErrInput, "price must be positive"))
	}
	return errs
}

func (m CancelOrderMsg) Validate() error {
	var errs error

	errs = errors.AppendField(errs, "Metadata", m.Metadata.Validate())
	errs = errors.AppendField(errs, "order id ", validID(m.OrderID))
	return errs
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
