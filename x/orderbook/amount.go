package orderbook

import (
	"encoding/binary"

	"github.com/iov-one/weave/coin"
	"github.com/iov-one/weave/errors"
)

// NewAmount takes whole and fractional, returns Amount struct
func NewAmount(whole int64, fractional int64) Amount {
	return Amount{
		Whole:      whole,
		Fractional: fractional,
	}
}

// NewAmountp takes whole and fractional, returns Amount struct pointer
func NewAmountp(whole int64, fractional int64) *Amount {
	return &Amount{
		Whole:      whole,
		Fractional: fractional,
	}
}

// Clone copies values of Amount to a new Amount struct
func (a *Amount) Clone() *Amount {
	return &Amount{
		Whole:      a.Whole,
		Fractional: a.Fractional,
	}
}

// Validate "borrowed" from coin.Coin
func (a *Amount) Validate() error {
	if a == nil {
		return errors.Wrap(errors.ErrEmpty, "amount")
	}
	if a.Whole < coin.MinInt || a.Whole > coin.MaxInt {
		return errors.Wrap(errors.ErrOverflow, "whole")
	}
	if a.Fractional < coin.MinFrac || a.Fractional > coin.MaxFrac {
		return errors.Wrap(errors.ErrOverflow, "fractional")
	}
	// make sure signs match
	if a.Whole != 0 && a.Fractional != 0 &&
		((a.Whole > 0) != (a.Fractional > 0)) {
		return errors.Wrap(errors.ErrState, "mismatched sign")
	}
	return nil
}

// IsPositive returns true if the value is greater than 0
func (a *Amount) IsPositive() bool {
	return a.Whole > 0 ||
		(a.Whole == 0 && a.Fractional > 0)
}

// IsNegative returns true if the value is less than 0
func (a *Amount) IsNegative() bool {
	return a.Whole < 0 ||
		(a.Whole == 0 && a.Fractional < 0)
}

// Lexographic produces a lexographic ordering, such than
//
//   A.Lexographic() < B.Lexographic == A < B
//
// This is a very nice trick to get clean range queries over sensible value ranges in a key-value store
// It is defined as 8-byte-bigendian(Whole) || 8-byte-bigendian(Fractional)
//
// Returns an error on nil or negative amounts (which would not be lexographically ordered)
func (a *Amount) Lexographic() ([]byte, error) {
	if a == nil {
		return nil, errors.Wrap(errors.ErrEmpty, "amount nil")
	}
	if a.IsNegative() {
		return nil, errors.Wrap(errors.ErrAmount, "negative")
	}
	res := make([]byte, 16)
	binary.BigEndian.PutUint64(res, uint64(a.Whole))
	binary.BigEndian.PutUint64(res[8:], uint64(a.Fractional))
	return res, nil
}
