package category

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	entity "github.com/hyoaru/itala-api/internal/features/category/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

func TestCreateCategory_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := &fakeCategoryRepository{
			create: func(ctx context.Context, userID string, category entity.Category) error {
				return nil
			},
		}

		useCase := NewCreateCategory(client)

		request := CreateCategoryRequest{
			UserID: "user-1",
			Name:   "Groceries",
			Type:   valueobject.TransactionTypeExpense,
		}

		got, err := useCase.Execute(context.Background(), request)
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if _, err := uuid.Parse(got.ID); err != nil {
			t.Errorf("parse id: %v", err)
		}

		if got.Name != "Groceries" {
			t.Errorf("got %v, want %v", got.Name, "Groceries")
		}

		if got.TransactionType != valueobject.TransactionTypeExpense {
			t.Errorf("got %v, want %v", got.TransactionType, valueobject.TransactionTypeExpense)
		}

		if got.CreatedAt.IsZero() {
			t.Error("got zero created_at, want non-zero")
		}

		if !got.CreatedAt.Equal(got.UpdatedAt) {
			t.Errorf("got %v, want %v", got.UpdatedAt, got.CreatedAt)
		}

		if got.DeletedAt != nil {
			t.Errorf("got %v, want %v", got.DeletedAt, nil)
		}
	})

	t.Run("trims name", func(t *testing.T) {
		client := &fakeCategoryRepository{
			create: func(ctx context.Context, userID string, category entity.Category) error {
				return nil
			},
		}

		useCase := NewCreateCategory(client)

		request := CreateCategoryRequest{
			UserID: "user-1",
			Name:   "  Groceries  ",
			Type:   valueobject.TransactionTypeExpense,
		}

		got, err := useCase.Execute(context.Background(), request)
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if got.Name != "Groceries" {
			t.Errorf("got %v, want %v", got.Name, "Groceries")
		}
	})

	t.Run("repository error", func(t *testing.T) {
		boom := errors.New("boom")

		client := &fakeCategoryRepository{
			create: func(ctx context.Context, userID string, category entity.Category) error {
				return boom
			},
		}

		useCase := NewCreateCategory(client)

		request := CreateCategoryRequest{
			UserID: "user-1",
			Name:   "Groceries",
			Type:   valueobject.TransactionTypeExpense,
		}

		got, err := useCase.Execute(context.Background(), request)

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}

		if got != (CreateCategoryResponse{}) {
			t.Errorf("got %v, want %v", got, CreateCategoryResponse{})
		}
	})

	t.Run("writes expected category", func(t *testing.T) {
		var capturedUserID string
		var capturedCategory entity.Category

		client := &fakeCategoryRepository{
			create: func(ctx context.Context, userID string, category entity.Category) error {
				capturedUserID = userID
				capturedCategory = category
				return nil
			},
		}

		useCase := NewCreateCategory(client)

		request := CreateCategoryRequest{
			UserID: "user-1",
			Name:   "  Groceries  ",
			Type:   valueobject.TransactionTypeExpense,
		}

		_, _ = useCase.Execute(context.Background(), request)

		if capturedUserID != "user-1" {
			t.Errorf("got %v, want %v", capturedUserID, "user-1")
		}

		if capturedCategory.ID == "" {
			t.Error("got empty id, want non-empty")
		}

		if capturedCategory.Name != "Groceries" {
			t.Errorf("got %v, want %v", capturedCategory.Name, "Groceries")
		}

		if capturedCategory.TransactionType != valueobject.TransactionTypeExpense {
			t.Errorf("got %v, want %v", capturedCategory.TransactionType, valueobject.TransactionTypeExpense)
		}

		if !capturedCategory.CreatedAt.Equal(capturedCategory.UpdatedAt) {
			t.Errorf("got %v, want %v", capturedCategory.UpdatedAt, capturedCategory.CreatedAt)
		}

		if capturedCategory.DeletedAt != nil {
			t.Errorf("got %v, want %v", capturedCategory.DeletedAt, nil)
		}
	})
}
