package orderbook

import (
	"github.com/iov-one/weave"
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

func (CreateOrderBookMsg) Validate() error {
	return nil
}

func (CreateOrderMsg) Validate() error {
	return nil
}

func (CancelOrderMsg) Validate() error {
	return nil
}
