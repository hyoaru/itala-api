package valueobject

import (
	"errors"

	"github.com/shopspring/decimal"
)

var ErrDivisionByZero = errors.New("division by zero")

type Decimal struct {
	value decimal.Decimal
}

func NewDecimal(value string) (Decimal, error) {
	amount, err := decimal.NewFromString(value)
	if err != nil {
		return Decimal{}, err
	}
	return Decimal{value: amount}, nil
}

func (d Decimal) Add(amount Decimal) Decimal {
	return Decimal{d.value.Add(amount.value)}
}

func (d Decimal) Subtract(amount Decimal) Decimal {
	return Decimal{d.value.Sub(amount.value)}
}

func (d Decimal) Negate() Decimal {
	return Decimal{d.value.Neg()}
}

func (d Decimal) Multiply(amount Decimal) Decimal {
	return Decimal{d.value.Mul(amount.value)}
}

func (d Decimal) Divide(amount Decimal) (Decimal, error) {
	if amount.value.IsZero() {
		return Decimal{}, ErrDivisionByZero
	}

	return Decimal{d.value.Div(amount.value)}, nil
}

func (d Decimal) String() string {
	return d.value.String()
}

func (d *Decimal) UnmarshalJSON(data []byte) error {
	value, err := NewDecimal(string(data))
	if err != nil {
		return err
	}

	*d = value
	return nil
}
