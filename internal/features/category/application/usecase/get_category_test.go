package category

import (
	"context"
	"errors"
	"testing"
	"time"

	entity "github.com/hyoaru/itala-api/internal/features/category/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

func TestGetCategory_Execute(t *testing.T) {
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
		}

		useCase := NewGetCategory(client)

		request := GetCategoryRequest{UserID: "user-1", ID: "category-1"}

		got, err := useCase.Execute(context.Background(), request)
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if want := GetCategoryResponse(category); got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		boom := errors.New("boom")

		client := &fakeCategoryRepository{
			findOne: func(ctx context.Context, userID string, categoryID string) (entity.Category, error) {
				return entity.Category{}, boom
			},
		}

		useCase := NewGetCategory(client)

		request := GetCategoryRequest{UserID: "user-1", ID: "category-1"}

		got, err := useCase.Execute(context.Background(), request)

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}

		if got != (GetCategoryResponse{}) {
			t.Errorf("got %v, want %v", got, GetCategoryResponse{})
		}
	})

	t.Run("writes expected request", func(t *testing.T) {
		var capturedUserID string
		var capturedID string

		client := &fakeCategoryRepository{
			findOne: func(ctx context.Context, userID string, categoryID string) (entity.Category, error) {
				capturedUserID = userID
				capturedID = categoryID
				return entity.Category{}, nil
			},
		}

		useCase := NewGetCategory(client)

		request := GetCategoryRequest{UserID: "user-1", ID: "category-1"}

		_, _ = useCase.Execute(context.Background(), request)

		if capturedUserID != "user-1" {
			t.Errorf("got %v, want %v", capturedUserID, "user-1")
		}

		if capturedID != "category-1" {
			t.Errorf("got %v, want %v", capturedID, "category-1")
		}
	})
}
