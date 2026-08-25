package account

import (
	"time"

	accountvalueobject "github.com/hyoaru/itala-api/internal/features/account/domain/valueobject"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

type Account struct {
	ID        string
	Name      string
	Balance   valueobject.Decimal
	Status    accountvalueobject.Status
	CreatedAt time.Time
	UpdatedAt time.Time
}
