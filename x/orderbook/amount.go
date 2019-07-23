package orderbook

import (
	"encoding/binary"

	"github.com/iov-one/weave/coin"
	"github.com/iov-one/weave/errors"
)

func NewAmount(whole int64, fractional int64) Amount {
	return Amount{
		Whole:      whole,
		Fractional: fractional,
	}
}

func NewAmountp(whole int64, fractional int64) *Amount {
	return &Amount{
		Whole:      whole,
		Fractional: fractional,
	}
}

func (a *Amount) Clone() *Amount {
	if a == nil {
		return nil
	}
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

var zeroAmount = new(Amount)

// Equals returns true if both amounts are equal
func (a *Amount) Equals(b *Amount) bool {
	return a.Whole == b.Whole && a.Fractional == b.Fractional
}

// Greater returns true if a > b
func (a *Amount) Greater(b *Amount) bool {
	return a.Whole > b.Whole ||
		(a.Whole == b.Whole && a.Fractional > b.Fractional)
}

// IsPositive returns true if the value is greater than 0
func (a *Amount) IsPositive() bool {
	return a.Greater(zeroAmount)
}

// IsNegative returns true if the value is less than 0
func (a *Amount) IsNegative() bool {
	return zeroAmount.Greater(a)
}

// Multiply returns a new coin of c multiplied by the decimal value of a
func (a *Amount) Multiply(c coin.Coin) (coin.Coin, error) {
	out, err := c.Multiply(a.Whole)
	if err != nil {
		return coin.Coin{}, err
	}

	frac, err := a.multiplyFraction(c)
	if err != nil {
		return coin.Coin{}, err
	}

	return out.Add(frac)
}

func (a *Amount) multiplyFraction(c coin.Coin) (coin.Coin, error) {
	var whole int64
	// TODO: do we need to check for overflow?
	fr := c.Whole*a.Fractional + (c.Fractional * a.Fractional / coin.FracUnit)
	if fr > coin.FracUnit {
		whole = fr / coin.FracUnit
		fr -= whole * coin.FracUnit
	}
	return coin.Coin{
		Whole:      whole,
		Fractional: fr,
		Ticker:     c.Ticker,
	}, nil
}

// Divide returns a new coin of c divided by the decimal value of a
func (a *Amount) Divide(c coin.Coin) (coin.Coin, error) {
	// TODO
	return coin.Coin{}, errors.Wrap(errors.ErrHuman, "not implemented")
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
