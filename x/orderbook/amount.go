package orderbook

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
