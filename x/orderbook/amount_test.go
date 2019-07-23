package orderbook

import (
	"bytes"
	"testing"

	"github.com/iov-one/weave/coin"
	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/weavetest/assert"
)

func TestAmountLexographicEncoding(t *testing.T) {
	cases := map[string]struct {
		a      Amount
		expect []byte
	}{
		"one byte encoding": {
			a:      Amount{Whole: 123, Fractional: 66},
			expect: []byte{0, 0, 0, 0, 0, 0, 0, 123, 0, 0, 0, 0, 0, 0, 0, 66},
		},
		"multi byte encoding": {
			a:      Amount{Whole: 0x7a4501, Fractional: 12345},
			expect: []byte{0, 0, 0, 0, 0, 0x7a, 0x45, 0x01, 0, 0, 0, 0, 0, 0, 0x30, 0x39},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			lex, err := tc.a.Lexographic()
			assert.Nil(t, err)
			assert.Equal(t, tc.expect, lex)
		})
	}
}

func TestAmountLexographic(t *testing.T) {
	cases := map[string]struct {
		a            Amount
		b            Amount
		expectBigger bool
	}{
		"one byte encoding": {
			a:            Amount{Whole: 123, Fractional: 66},
			b:            Amount{Whole: 123, Fractional: 270},
			expectBigger: false,
		},
		"multi byte encoding": {
			a:            Amount{Whole: 260, Fractional: 66},
			b:            Amount{Whole: 123, Fractional: 270},
			expectBigger: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			a, err := tc.a.Lexographic()
			assert.Nil(t, err)
			b, err := tc.b.Lexographic()
			assert.Nil(t, err)
			cmp := bytes.Compare(a, b)
			if tc.expectBigger {
				if cmp <= 0 {
					t.Fatalf("Expected a to be larger")
				}
			} else {
				if cmp >= 0 {
					t.Fatalf("Expected a to be smaller")
				}
			}
		})
	}
}

func TestAmountGreater(t *testing.T) {
	cases := map[string]struct {
		a        Amount
		b        Amount
		expected bool
	}{
		"fractional diff, lower": {
			a:        Amount{Whole: 123, Fractional: 66},
			b:        Amount{Whole: 123, Fractional: 270},
			expected: false,
		},
		"fractional diff, higher": {
			a:        Amount{Whole: 123, Fractional: 270},
			b:        Amount{Whole: 123, Fractional: 88},
			expected: true,
		},
		"fractional diff, equal": {
			a:        Amount{Whole: 123, Fractional: 267},
			b:        Amount{Whole: 123, Fractional: 267},
			expected: false,
		},
		"whole diff, greater": {
			a:        Amount{Whole: 187, Fractional: 267},
			b:        Amount{Whole: 123, Fractional: 267},
			expected: true,
		},
		"whole diff, lower, fractional greater": {
			a:        Amount{Whole: 187, Fractional: 267},
			b:        Amount{Whole: 188, Fractional: 0},
			expected: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			greater := tc.a.Greater(&tc.b)
			assert.Equal(t, tc.expected, greater)
		})
	}
}

func TestAmountMultiply(t *testing.T) {
	cases := map[string]struct {
		in       coin.Coin
		amt      Amount
		expected coin.Coin
		wantErr  *errors.Error
	}{
		"multiply by 1": {
			in:       coin.NewCoin(100, 200, "ETH"),
			amt:      NewAmount(1, 0),
			expected: coin.NewCoin(100, 200, "ETH"),
		},
		"multiply by 0": {
			in:       coin.NewCoin(100, 200, "ETH"),
			amt:      NewAmount(0, 0),
			expected: coin.NewCoin(0, 0, "ETH"),
		},
		"multiply by positive int": {
			in:       coin.NewCoin(123, 456, "ETH"),
			amt:      NewAmount(17, 0),
			expected: coin.NewCoin(2091, 7752, "ETH"),
		},
		"multiply by int, fractional to whole overflow": {
			in:  coin.NewCoin(20, 100000000, "ETH"),
			amt: NewAmount(12, 0),
			// 1 rolls on over into whole
			expected: coin.NewCoin(241, 200000000, "ETH"),
		},
		"multiply by fractional 0.1, no underflow": {
			in:       coin.NewCoin(20, 100000000, "ATM"),
			amt:      NewAmount(0, 100000000),
			expected: coin.NewCoin(2, 10000000, "ATM"),
		},
		"multiply by fractional 0.1,  underflow": {
			in:       coin.NewCoin(12, 345000000, "ATM"),
			amt:      NewAmount(0, 100000000),
			expected: coin.NewCoin(1, 234500000, "ATM"),
		},
		"multiply by fractional 2.3, addition": {
			in:       coin.NewCoin(12, 345000000, "ATM"),
			amt:      NewAmount(2, 300000000),
			expected: coin.NewCoin(28, 393500000, "ATM"),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Nil(t, tc.in.Validate())
			out, err := tc.amt.Multiply(tc.in)

			if tc.wantErr != nil {
				if !tc.wantErr.Is(err) {
					t.Fatalf("Unexpected error: %v", err)
				}
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, tc.expected, out)
		})
	}
}
