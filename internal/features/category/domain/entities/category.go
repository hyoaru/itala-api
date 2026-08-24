package category

import (
	"time"

	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
)

type Category struct {
	ID              string
	Name            string
	TransactionType valueobjects.TransactionType
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
