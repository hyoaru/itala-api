package category

import (
	"time"

	categoryvo "github.com/hyoaru/itala-api/internal/features/category/domain/valueobject"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

type Category struct {
	ID              string
	Name            string
	TransactionType valueobject.TransactionType
	Status          categoryvo.Status
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
