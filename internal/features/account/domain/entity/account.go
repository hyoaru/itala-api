package account

import (
	"time"

	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

type Account struct {
	ID        string
	Name      string
	Balance   valueobject.Decimal
	DeletedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
