package orderbook

import (
	"regexp"

	"github.com/iov-one/tutorial/morm"
	"github.com/iov-one/weave/coin"
	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/orm"
)

var _ morm.Model = (*Market)(nil)

var validMarketName = regexp.MustCompile(`^[a-zA-Z0-9_.-]{4,32}$`).MatchString

// SetID is a minimal implementation, useful when the ID is a separate protobuf field
func (m *Market) SetID(id []byte) error {
	m.ID = id
	return nil
}

// Copy produces a new copy to fulfill the Model interface
func (m *Market) Copy() orm.CloneableData {
	return &Market{
		Metadata: m.Metadata.Copy(),
		ID:       copyBytes(m.ID),
		Owner:    m.Owner.Clone(),
		Name:     m.Name,
	}
}

// Validate is always succesful
func (m *Market) Validate() error {
	if err := m.Metadata.Validate(); err != nil {
		return err
	}
	if err := isGenID(m.ID, true); err != nil {
		return err
	}
	if err := m.Owner.Validate(); err != nil {
		return errors.Wrap(err, "owner")
	}
	if !validMarketName(m.Name) {
		return errors.Wrap(errors.ErrModel, "invalid market name")
	}
	return nil
}

var _ morm.Model = (*OrderBook)(nil)

// SetID is a minimal implementation, useful when the ID is a separate protobuf field
func (o *OrderBook) SetID(id []byte) error {
	o.ID = id
	return nil
}

// Copy produces a new copy to fulfill the Model interface
func (o *OrderBook) Copy() orm.CloneableData {
	return &OrderBook{
		Metadata:      o.Metadata.Copy(),
		ID:            copyBytes(o.ID),
		MarketID:      copyBytes(o.MarketID),
		AskTicker:     o.AskTicker,
		BidTicker:     o.BidTicker,
		TotalAskCount: o.TotalAskCount,
		TotalBidCount: o.TotalBidCount,
	}
}

// Validate is always succesful
func (o *OrderBook) Validate() error {
	if err := o.Metadata.Validate(); err != nil {
		return err
	}
	if err := isGenID(o.ID, true); err != nil {
		return err
	}
	if err := isGenID(o.MarketID, false); err != nil {
		return errors.Wrap(err, "market id")
	}
	if !coin.IsCC(o.AskTicker) {
		return errors.Wrap(errors.ErrModel, "invalid ask ticker")
	}
	if !coin.IsCC(o.BidTicker) {
		return errors.Wrap(errors.ErrModel, "invalid bid ticker")
	}
	if o.TotalAskCount < 0 {
		return errors.Wrap(errors.ErrModel, "negative total ask count")
	}
	if o.TotalBidCount < 0 {
		return errors.Wrap(errors.ErrModel, "negative total bid count")
	}
	return nil
}

var _ morm.Model = (*Order)(nil)

// SetID is a minimal implementation, useful when the ID is a separate protobuf field
func (o *Order) SetID(id []byte) error {
	o.ID = id
	return nil
}

// Copy produces a new copy to fulfill the Model interface
func (o *Order) Copy() orm.CloneableData {
	return &Order{
		Metadata:       o.Metadata.Copy(),
		ID:             copyBytes(o.ID),
		Trader:         copyBytes(o.Trader),
		OrderBookID:    copyBytes(o.OrderBookID),
		Side:           o.Side,
		OrderState:     o.OrderState,
		OriginalOffer:  o.OriginalOffer.Clone(),
		RemainingOffer: o.RemainingOffer.Clone(),
		Price:          o.Price.Clone(),
		TradeIds:       copyBytesList(o.TradeIds),
		CreatedAt:      o.CreatedAt,
		UpdatedAt:      o.UpdatedAt,
	}
}

// Validate is always succesful
func (o *Order) Validate() error {
	if err := o.Metadata.Validate(); err != nil {
		return err
	}
	if err := isGenID(o.ID, true); err != nil {
		return err
	}
	if err := o.Trader.Validate(); err != nil {
		return errors.Wrap(err, "trader")
	}
	if err := isGenID(o.OrderBookID, false); err != nil {
		return errors.Wrap(err, "order book id")
	}
	// TODO: valid Side?
	// TODO: valid OrderState?
	if o.OriginalOffer == nil {
		return errors.Wrap(errors.ErrEmpty, "original offer")
	}
	if err := o.OriginalOffer.Validate(); err != nil {
		return errors.Wrap(err, "original offer")
	}
	if o.RemainingOffer == nil {
		return errors.Wrap(errors.ErrEmpty, "remaining offer")
	}
	if err := o.RemainingOffer.Validate(); err != nil {
		return errors.Wrap(err, "remaining offer")
	}

	// TODO: valid price
	// TODO: valid trade ids (also rethink how we handle this? just use index and not in model?)

	if err := o.UpdatedAt.Validate(); err != nil {
		return errors.Wrap(err, "updated at")
	}
	if o.UpdatedAt == 0 {
		return errors.Wrap(errors.ErrEmpty, "missing updated at")
	}
	if err := o.CreatedAt.Validate(); err != nil {
		return errors.Wrap(err, "created at")
	}
	if o.CreatedAt == 0 {
		return errors.Wrap(errors.ErrEmpty, "missing created at")
	}
	return nil
}

var _ morm.Model = (*Trade)(nil)

// SetID is a minimal implementation, useful when the ID is a separate protobuf field
func (t *Trade) SetID(id []byte) error {
	t.ID = id
	return nil
}

// Copy produces a new copy to fulfill the Model interface
func (t *Trade) Copy() orm.CloneableData {
	return &Trade{
		Metadata:    t.Metadata.Copy(),
		ID:          copyBytes(t.ID),
		OrderBookID: copyBytes(t.OrderBookID),
		OrderID:     copyBytes(t.OrderID),
		Taker:       copyBytes(t.Taker),
		Maker:       copyBytes(t.Maker),
		MakerPaid:   t.MakerPaid.Clone(),
		TakerPaid:   t.TakerPaid.Clone(),
		ExecutedAt:  t.ExecutedAt,
	}
}

// Validate is always succesful
func (o *Trade) Validate() error {
	if err := o.Metadata.Validate(); err != nil {
		return err
	}
	if err := isGenID(o.ID, true); err != nil {
		return err
	}
	if err := isGenID(o.OrderBookID, false); err != nil {
		return errors.Wrap(err, "order book id")
	}
	if err := isGenID(o.OrderID, false); err != nil {
		return errors.Wrap(err, "order  id")
	}
	if err := o.Taker.Validate(); err != nil {
		return errors.Wrap(err, "taker")
	}
	if err := o.Maker.Validate(); err != nil {
		return errors.Wrap(err, "maker")
	}
	if o.MakerPaid == nil {
		return errors.Wrap(errors.ErrEmpty, "maker paid")
	}
	if err := o.MakerPaid.Validate(); err != nil {
		return errors.Wrap(err, "maker paid")
	}
	if o.TakerPaid == nil {
		return errors.Wrap(errors.ErrEmpty, "maker paid")
	}
	if err := o.TakerPaid.Validate(); err != nil {
		return errors.Wrap(err, "taker paid")
	}
	if err := o.ExecutedAt.Validate(); err != nil {
		return errors.Wrap(err, "executed at")
	}
	if o.ExecutedAt == 0 {
		return errors.Wrap(errors.ErrEmpty, "missing executed at")
	}
	return nil
}

// isGenID ensures that the ID is 8 byte input.
// if allowEmpty is set, we also allow empty
func isGenID(id []byte, allowEmpty bool) error {
	if len(id) == 0 {
		if allowEmpty {
			return nil
		}
		return errors.Wrap(errors.ErrEmpty, "missing id")
	}
	if len(id) != 8 {
		return errors.Wrap(errors.ErrInput, "id must be 8 bytes")
	}
	return nil
}

func copyBytes(in []byte) []byte {
	if in == nil {
		return nil
	}
	cpy := make([]byte, len(in))
	copy(cpy, in)
	return cpy
}

func copyBytesList(in [][]byte) [][]byte {
	if in == nil {
		return nil
	}
	cpy := make([][]byte, len(in))
	for i, bz := range in {
		cpy[i] = copyBytes(bz)
	}
	return cpy
}
