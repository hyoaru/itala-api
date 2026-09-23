package category

import (
	"context"
	"errors"
	"testing"
)

func TestDeleteCategory_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := &fakeCategoryRepository{
			delete: func(ctx context.Context, userID string, categoryID string) error {
				return nil
			},
		}

		useCase := NewDeleteCategory(client)

		request := DeleteCategoryRequest{UserID: "user-1", ID: "category-1"}

		if _, err := useCase.Execute(context.Background(), request); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		boom := errors.New("boom")

		client := &fakeCategoryRepository{
			delete: func(ctx context.Context, userID string, categoryID string) error {
				return boom
			},
		}

		useCase := NewDeleteCategory(client)

		request := DeleteCategoryRequest{UserID: "user-1", ID: "category-1"}

		_, err := useCase.Execute(context.Background(), request)

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}
	})

	t.Run("writes expected request", func(t *testing.T) {
		var capturedUserID string
		var capturedID string

		client := &fakeCategoryRepository{
			delete: func(ctx context.Context, userID string, categoryID string) error {
				capturedUserID = userID
				capturedID = categoryID
				return nil
			},
		}

		useCase := NewDeleteCategory(client)

		request := DeleteCategoryRequest{UserID: "user-1", ID: "category-1"}

		_, _ = useCase.Execute(context.Background(), request)

		if capturedUserID != "user-1" {
			t.Errorf("got %v, want %v", capturedUserID, "user-1")
		}

		if capturedID != "category-1" {
			t.Errorf("got %v, want %v", capturedID, "category-1")
		}
	})
}
