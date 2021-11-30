package domain

type Exchange struct {
	Symbol
	Period
	Amount
}

func (e Exchange) IsValid() bool {
	return e.Symbol.IsValid() && e.Period.IsValid()
}
