package category

import (
	"time"

	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

type Category struct {
	ID              string
	Name            string
	TransactionType valueobject.TransactionType
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
