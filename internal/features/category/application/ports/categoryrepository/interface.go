package category

import (
	"context"

	entities "github.com/hyoaru/itala-api/internal/features/category/domain/entities"
)

type CategoryRepository interface {
	Create(ctx context.Context, userID string, category entities.Category) error
}
