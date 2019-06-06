package orderbook

import (
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
