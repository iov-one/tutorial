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
	var errs error

	//errs = errors.AppendField(errs, "Metadata", m.Metadata.Validate())
	errs = errors.AppendField(errs, "ID", isGenID(m.ID, true))
	errs = errors.AppendField(errs, "Owner", m.Owner.Validate())

	if !validMarketName(m.Name) {
		errs = errors.AppendField(errs, "MarketName", errors.ErrModel)
	}

	return errs
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
	var errs error

	//errs = errors.AppendField(errs, "Metadata", o.Metadata.Validate())
	errs = errors.AppendField(errs, "ID", isGenID(o.ID, true))
	errs = errors.AppendField(errs, "MarketID", isGenID(o.MarketID, false))

	if !coin.IsCC(o.AskTicker) {
		errs = errors.AppendField(errs, "AskTicker", errors.ErrCurrency)
	}
	if !coin.IsCC(o.BidTicker) {
		errs = errors.AppendField(errs, "BidTicker", errors.ErrCurrency)
	}

	if o.TotalAskCount < 0 {
		errs = errors.AppendField(errs, "TotalAskCount", errors.ErrModel)
	}
	if o.TotalBidCount < 0 {
		errs = errors.AppendField(errs, "TotalBidCount", errors.ErrModel)
	}

	return errs
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
	var errs error

	//errs = errors.AppendField(errs, "Metadata", o.Metadata.Validate())
	errs = errors.AppendField(errs, "ID", isGenID(o.ID, true))
	errs = errors.AppendField(errs, "Trader", o.Trader.Validate())
	errs = errors.AppendField(errs, "OrderBookID", isGenID(o.OrderBookID, false))

	if o.Side != Side_Ask && o.Side != Side_Bid {
		errs = errors.AppendField(errs, "Side", errors.ErrState)
	}
	if o.OrderState != OrderState_Open && o.OrderState != OrderState_Done && o.OrderState != OrderState_Cancel {
		errs = errors.AppendField(errs, "OrderState", errors.ErrState)
	}

	if o.OriginalOffer == nil {
		errs = errors.AppendField(errs, "OriginalOffer", errors.ErrEmpty)
	} else if err := o.OriginalOffer.Validate(); err != nil {
		errs = errors.AppendField(errs, "OriginalOffer", err)
	}
	if o.RemainingOffer == nil {
		errs = errors.AppendField(errs, "RemainingOffer", errors.ErrEmpty)
	} else if err := o.RemainingOffer.Validate(); err != nil {
		errs = errors.AppendField(errs, "RemaningOffer", err)
	}

	if err := o.Price.Validate(); err != nil {
		errs = errors.AppendField(errs, "Price", o.Price.Validate())
	} else if !o.Price.IsPositive() {
		errs = errors.Append(errs,
			errors.Field("Price", errors.ErrState, "price must be positive"))
	}
	// TODO: valid trade ids (also rethink how we handle this? just use index and not in model?)

	if err := o.UpdatedAt.Validate(); err != nil {
		errs = errors.AppendField(errs, "UpdatedAt", o.UpdatedAt.Validate())
	} else if o.UpdatedAt == 0 {
		errs = errors.AppendField(errs, "UpdatedAt", errors.ErrEmpty)
	}

	if err := o.CreatedAt.Validate(); err != nil {
		errs = errors.AppendField(errs, "CreatedAt", o.CreatedAt.Validate())
	} else if o.CreatedAt == 0 {
		errs = errors.AppendField(errs, "CreatedAt", errors.ErrEmpty)
	}

	return errs
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
func (t *Trade) Validate() error {
	var errs error

	//errs = errors.AppendField(errs, "metadata", t.Metadata.Validate())

	errs = errors.AppendField(errs, "ID", isGenID(t.ID, true))
	errs = errors.AppendField(errs, "OrderBookID", isGenID(t.OrderBookID, false))
	errs = errors.AppendField(errs, "OrderID", isGenID(t.OrderID, false))
	errs = errors.AppendField(errs, "Taker", t.Taker.Validate())
	errs = errors.AppendField(errs, "Maker", t.Maker.Validate())

	if t.MakerPaid == nil {
		errs = errors.AppendField(errs, "MakerPaid", errors.ErrEmpty)
	} else if err := t.MakerPaid.Validate(); err != nil {
		errs = errors.AppendField(errs, "MakerPaid", err)
	}

	if t.TakerPaid == nil {
		errs = errors.AppendField(errs, "TakerPaid", errors.ErrEmpty)
	} else if err := t.TakerPaid.Validate(); err != nil {
		errs = errors.AppendField(errs, "TakerPaid", err)
	}

	errs = errors.AppendField(errs, "ExecutedAt", t.ExecutedAt.Validate())
	if err := t.ExecutedAt.Validate(); err != nil {
		errors.AppendField(errs, "ExecutedAt", t.ExecutedAt.Validate())
	} else if t.ExecutedAt == 0 {
		errs = errors.AppendField(errs, "ExecutedAt", errors.ErrEmpty)
	}

	return errs
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
