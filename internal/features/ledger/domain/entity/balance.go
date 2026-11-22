package entity

import "github.com/hyoaru/itala-api/internal/shared/domain/valueobject"

type BalanceDelta struct {
	AccountID string
	Amount    valueobject.Decimal
}
