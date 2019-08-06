package orderbook

import (
	"testing"
	"time"

	"github.com/iov-one/weave"
	"github.com/iov-one/weave/coin"
	"github.com/iov-one/weave/errors"
	"github.com/iov-one/weave/weavetest"
	"github.com/iov-one/weave/weavetest/assert"

	"github.com/iov-one/tutorial/morm"
)

func TestValidateOrderBook(t *testing.T) {
	cases := map[string]struct {
		model    morm.Model
		wantErrs map[string]*errors.Error
	}{
		"success, no id": {
			model: &OrderBook{
				Metadata:  &weave.Metadata{Schema: 1},
				MarketID:  weavetest.SequenceID(5),
				AskTicker: "BAR",
				BidTicker: "FOO",
			},
			wantErrs: map[string]*errors.Error{
				"ID":            nil,
				"Metadata":      nil,
				"MarketID":      nil,
				"AskTicker":     nil,
				"BidTicker":     nil,
				"TotalAskCount": nil,
				"TotalBidCount": nil,
			},
		},
		"success, with id": {
			model: &OrderBook{
				Metadata:  &weave.Metadata{Schema: 1},
				ID:        weavetest.SequenceID(13),
				MarketID:  weavetest.SequenceID(1),
				AskTicker: "BAR",
				BidTicker: "FOO",
			},
			wantErrs: map[string]*errors.Error{
				"ID":            nil,
				"Metadata":      nil,
				"MarketID":      nil,
				"AskTicker":     nil,
				"BidTicker":     nil,
				"TotalAskCount": nil,
				"TotalBidCount": nil,
			},
		},
		"failure, no market id": {
			model: &OrderBook{
				Metadata:  &weave.Metadata{Schema: 1},
				ID:        weavetest.SequenceID(13),
				AskTicker: "BAR",
				BidTicker: "FOO",
			},
			wantErrs: map[string]*errors.Error{
				"ID":            nil,
				"Metadata":      nil,
				"MarketID":      errors.ErrEmpty,
				"AskTicker":     nil,
				"BidTicker":     nil,
				"TotalAskCount": nil,
				"TotalBidCount": nil,
			},
		},
	}
	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			err := tc.model.Validate()
			for field, wantErr := range tc.wantErrs {
				assert.FieldError(t, err, field, wantErr)
			}
		})
	}
}

func TestValidateMarket(t *testing.T) {
	cases := map[string]struct {
		model    morm.Model
		wantErrs map[string]*errors.Error
	}{
		"success, no id": {
			model: &Market{
				Metadata: &weave.Metadata{Schema: 1},
				Owner:    weavetest.NewCondition().Address(),
				Name:     "Fred",
			},
			wantErrs: map[string]*errors.Error{
				"ID":         nil,
				"Owner":      nil,
				"MarketName": nil,
			},
		},
		"success, with id": {
			model: &Market{
				Metadata: &weave.Metadata{Schema: 1},
				ID:       weavetest.SequenceID(2),
				Owner:    weavetest.NewCondition().Address(),
				Name:     "Fred",
			},
			wantErrs: map[string]*errors.Error{
				"ID":         nil,
				"Owner":      nil,
				"MarketName": nil,
			},
		},
		"failure bad owner": {
			model: &Market{
				Metadata: &weave.Metadata{Schema: 1},
				ID:       weavetest.SequenceID(2),
				Owner:    []byte("foobar"),
				Name:     "Fred",
			},
			wantErrs: map[string]*errors.Error{
				"ID":         nil,
				"Owner":      errors.ErrInput,
				"MarketName": nil,
			},
		},
	}
	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			err := tc.model.Validate()
			for field, wantErr := range tc.wantErrs {
				assert.FieldError(t, err, field, wantErr)
			}
		})
	}
}

func TestValidateOrder(t *testing.T) {
	now := weave.AsUnixTime(time.Now())

	cases := map[string]struct {
		model    morm.Model
		wantErrs map[string]*errors.Error
	}{
		"success, no id": {
			model: &Order{
				Metadata:       &weave.Metadata{Schema: 1},
				Trader:         weavetest.NewCondition().Address(),
				OrderBookID:    weavetest.SequenceID(5),
				Side:           Side_Ask,
				OrderState:     OrderState_Open,
				OriginalOffer:  coin.NewCoinp(100, 0, "ETH"),
				RemainingOffer: coin.NewCoinp(50, 17, "ETH"),
				Price:          NewAmountp(121, 0),
				CreatedAt:      now,
				UpdatedAt:      now,
			},
			wantErrs: map[string]*errors.Error{
				"ID":             nil,
				"Trader":         nil,
				"OrderBookID":    nil,
				"Side":           nil,
				"OrderState":     nil,
				"OriginalOffer":  nil,
				"RemainingOffer": nil,
				"Price":          nil,
				"UpdatedAt":      nil,
				"CreatedAt":      nil,
			},
		},
		"success, with id": {
			model: &Order{
				Metadata:       &weave.Metadata{Schema: 1},
				ID:             weavetest.SequenceID(17),
				Trader:         weavetest.NewCondition().Address(),
				OrderBookID:    weavetest.SequenceID(5),
				Side:           Side_Ask,
				OrderState:     OrderState_Open,
				OriginalOffer:  coin.NewCoinp(100, 0, "ETH"),
				RemainingOffer: coin.NewCoinp(50, 17, "ETH"),
				Price:          NewAmountp(121, 0),
				CreatedAt:      now,
				UpdatedAt:      now,
			},
			wantErrs: map[string]*errors.Error{
				"ID":             nil,
				"Trader":         nil,
				"OrderBookID":    nil,
				"Side":           nil,
				"OrderState":     nil,
				"OriginalOffer":  nil,
				"RemainingOffer": nil,
				"Price":          nil,
				"UpdatedAt":      nil,
				"CreatedAt":      nil,
			},
		},
		"missing timestamps": {
			model: &Order{
				Metadata:       &weave.Metadata{Schema: 1},
				ID:             weavetest.SequenceID(17),
				Trader:         weavetest.NewCondition().Address(),
				OrderBookID:    weavetest.SequenceID(5),
				Side:           Side_Ask,
				OrderState:     OrderState_Open,
				OriginalOffer:  coin.NewCoinp(100, 0, "ETH"),
				RemainingOffer: coin.NewCoinp(50, 17, "ETH"),
				Price:          NewAmountp(121, 0),
			},
			wantErrs: map[string]*errors.Error{
				"ID":             nil,
				"Trader":         nil,
				"OrderBookID":    nil,
				"Side":           nil,
				"OrderState":     nil,
				"OriginalOffer":  nil,
				"RemainingOffer": nil,
				"Price":          nil,
				"UpdatedAt":      errors.ErrEmpty,
				"CreatedAt":      errors.ErrEmpty,
			},
		},
	}
	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			err := tc.model.Validate()
			for field, wantErr := range tc.wantErrs {
				assert.FieldError(t, err, field, wantErr)
			}
		})
	}
}

func TestValidateTrade(t *testing.T) {
	now := weave.AsUnixTime(time.Now())

	cases := map[string]struct {
		model    morm.Model
		wantErrs map[string]*errors.Error
	}{
		"success, no id": {
			model: &Trade{
				Metadata:    &weave.Metadata{Schema: 1},
				OrderID:     weavetest.SequenceID(14),
				OrderBookID: weavetest.SequenceID(2),
				Taker:       weavetest.NewCondition().Address(),
				Maker:       weavetest.NewCondition().Address(),
				TakerPaid:   coin.NewCoinp(100, 0, "ETH"),
				MakerPaid:   coin.NewCoinp(7, 234456, "BTC"),
				ExecutedAt:  now,
			},
			wantErrs: map[string]*errors.Error{
				"ID":          nil,
				"OrderBookID": nil,
				"OrderID":     nil,
				"Taker":       nil,
				"Maker":       nil,
				"TakerPaid":   nil,
				"MakerPaid":   nil,
				"ExecutedAt":  nil,
			},
		},
		"success, with id": {
			model: &Trade{
				Metadata:    &weave.Metadata{Schema: 1},
				ID:          weavetest.SequenceID(7654),
				OrderID:     weavetest.SequenceID(14),
				OrderBookID: weavetest.SequenceID(2),
				Taker:       weavetest.NewCondition().Address(),
				Maker:       weavetest.NewCondition().Address(),
				TakerPaid:   coin.NewCoinp(100, 0, "ETH"),
				MakerPaid:   coin.NewCoinp(7, 234456, "BTC"),
				ExecutedAt:  now,
			},
			wantErrs: map[string]*errors.Error{
				"ID":          nil,
				"OrderBookID": nil,
				"OrderID":     nil,
				"Taker":       nil,
				"Maker":       nil,
				"TakerPaid":   nil,
				"MakerPaid":   nil,
				"ExecutedAt":  nil,
			},
		},
		"missing payment": {
			model: &Trade{
				Metadata:    &weave.Metadata{Schema: 1},
				ID:          weavetest.SequenceID(7654),
				OrderID:     weavetest.SequenceID(14),
				OrderBookID: weavetest.SequenceID(2),
				Taker:       weavetest.NewCondition().Address(),
				Maker:       weavetest.NewCondition().Address(),
				ExecutedAt:  now,
			},
			wantErrs: map[string]*errors.Error{
				"ID":          nil,
				"OrderBookID": nil,
				"OrderID":     nil,
				"Taker":       nil,
				"Maker":       nil,
				"TakerPaid":   errors.ErrEmpty,
				"MakerPaid":   errors.ErrEmpty,
				"ExecutedAt":  nil,
			},
		},
	}
	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			err := tc.model.Validate()
			for field, wantErr := range tc.wantErrs {
				assert.FieldError(t, err, field, wantErr)
			}
		})
	}
}
