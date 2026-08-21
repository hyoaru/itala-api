package account

import (
	"time"

	"github.com/shopspring/decimal"
)

type Account struct {
	ID        string
	Name      string
	Balance   decimal.Decimal
	CreatedAt time.Time
	UpdatedAt time.Time
}
