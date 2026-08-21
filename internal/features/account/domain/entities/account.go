package account

import (
	"time"

	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
)

type Account struct {
	ID        string
	Name      string
	Balance   valueobjects.Decimal
	CreatedAt time.Time
	UpdatedAt time.Time
}
