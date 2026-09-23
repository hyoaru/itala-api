package category

import (
	"context"
	"errors"
	"testing"
	"time"

	entity "github.com/hyoaru/itala-api/internal/features/category/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

func TestUpdateCategory_Execute(t *testing.T) {
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	category := entity.Category{
		ID:              "category-1",
		Name:            "Groceries",
		TransactionType: valueobject.TransactionTypeExpense,
		CreatedAt:       createdAt,
		UpdatedAt:       createdAt,
	}

	t.Run("success", func(t *testing.T) {
		client := &fakeCategoryRepository{
			findOne: func(ctx context.Context, userID string, categoryID string) (entity.Category, error) {
				return category, nil
			},
			update: func(ctx context.Context, userID string, category entity.Category) error {
				return nil
			},
		}

		useCase := NewUpdateCategory(client)

		request := UpdateCategoryRequest{UserID: "user-1", ID: "category-1", Name: "Groceries"}

		if _, err := useCase.Execute(context.Background(), request); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}
	})

	t.Run("get current error", func(t *testing.T) {
		boom := errors.New("boom")

		client := &fakeCategoryRepository{
			findOne: func(ctx context.Context, userID string, categoryID string) (entity.Category, error) {
				return entity.Category{}, boom
			},
		}

		useCase := NewUpdateCategory(client)

		request := UpdateCategoryRequest{UserID: "user-1", ID: "category-1", Name: "Groceries"}

		_, err := useCase.Execute(context.Background(), request)

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		boom := errors.New("boom")

		client := &fakeCategoryRepository{
			findOne: func(ctx context.Context, userID string, categoryID string) (entity.Category, error) {
				return category, nil
			},
			update: func(ctx context.Context, userID string, category entity.Category) error {
				return boom
			},
		}

		useCase := NewUpdateCategory(client)

		request := UpdateCategoryRequest{UserID: "user-1", ID: "category-1", Name: "Groceries"}

		_, err := useCase.Execute(context.Background(), request)

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}
	})

	t.Run("writes expected category", func(t *testing.T) {
		var capturedUserID string
		var capturedCategory entity.Category

		client := &fakeCategoryRepository{
			findOne: func(ctx context.Context, userID string, categoryID string) (entity.Category, error) {
				return category, nil
			},
			update: func(ctx context.Context, userID string, category entity.Category) error {
				capturedUserID = userID
				capturedCategory = category
				return nil
			},
		}

		useCase := NewUpdateCategory(client)

		request := UpdateCategoryRequest{UserID: "user-1", ID: "category-1", Name: "  Groceries  "}

		_, _ = useCase.Execute(context.Background(), request)

		if capturedUserID != "user-1" {
			t.Errorf("got %v, want %v", capturedUserID, "user-1")
		}

		if capturedCategory.ID != "category-1" {
			t.Errorf("got %v, want %v", capturedCategory.ID, "category-1")
		}

		if capturedCategory.Name != "Groceries" {
			t.Errorf("got %v, want %v", capturedCategory.Name, "Groceries")
		}

		if capturedCategory.UpdatedAt.IsZero() {
			t.Error("got zero updated_at, want non-zero")
		}
	})
}
