package orderbook

import (
	"bytes"
	"testing"

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
