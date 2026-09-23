package category

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	categoryrepository "github.com/hyoaru/itala-api/internal/features/category/application/port/categoryrepository"
	entity "github.com/hyoaru/itala-api/internal/features/category/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

func TestListCategories_Execute(t *testing.T) {
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	category := entity.Category{
		ID:              "category-1",
		Name:            "Groceries",
		TransactionType: valueobject.TransactionTypeExpense,
		CreatedAt:       createdAt,
		UpdatedAt:       createdAt,
	}
	nextCursor := "cursor-1"

	t.Run("success", func(t *testing.T) {
		client := &fakeCategoryRepository{
			find: func(
				ctx context.Context,
				userID string,
				query categoryrepository.CategoryQuery,
			) (categoryrepository.CategoryPage, error) {
				return categoryrepository.CategoryPage{
					Categories: []entity.Category{category},
					NextCursor: &nextCursor,
				}, nil
			},
		}

		useCase := NewListCategories(client)

		request := ListCategoriesRequest{UserID: "user-1", Limit: 10}

		got, err := useCase.Execute(context.Background(), request)
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		want := ListCategoriesResponse{
			Categories: []entity.Category{category},
			NextCursor: &nextCursor,
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		boom := errors.New("boom")

		client := &fakeCategoryRepository{
			find: func(
				ctx context.Context,
				userID string,
				query categoryrepository.CategoryQuery,
			) (categoryrepository.CategoryPage, error) {
				return categoryrepository.CategoryPage{}, boom
			},
		}

		useCase := NewListCategories(client)

		request := ListCategoriesRequest{UserID: "user-1", Limit: 10}

		got, err := useCase.Execute(context.Background(), request)

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}

		if len(got.Categories) != 0 || got.NextCursor != nil {
			t.Errorf("got %v, want %v", got, ListCategoriesResponse{})
		}
	})

	t.Run("writes expected query", func(t *testing.T) {
		var capturedUserID string
		var capturedQuery categoryrepository.CategoryQuery

		client := &fakeCategoryRepository{
			find: func(
				ctx context.Context,
				userID string,
				query categoryrepository.CategoryQuery,
			) (categoryrepository.CategoryPage, error) {
				capturedUserID = userID
				capturedQuery = query
				return categoryrepository.CategoryPage{}, nil
			},
		}

		useCase := NewListCategories(client)

		name := "Groceries"
		transactionType := valueobject.TransactionTypeExpense
		cursor := "cursor-1"

		request := ListCategoriesRequest{
			UserID:          "user-1",
			Limit:           10,
			Name:            &name,
			TransactionType: &transactionType,
			Cursor:          &cursor,
		}

		_, _ = useCase.Execute(context.Background(), request)

		if capturedUserID != "user-1" {
			t.Errorf("got %v, want %v", capturedUserID, "user-1")
		}

		if capturedQuery.Limit != 10 {
			t.Errorf("got %v, want %v", capturedQuery.Limit, 10)
		}

		if capturedQuery.Name == nil || *capturedQuery.Name != "Groceries" {
			t.Errorf("got %v, want %v", capturedQuery.Name, "Groceries")
		}

		if capturedQuery.TransactionType == nil || *capturedQuery.TransactionType != valueobject.TransactionTypeExpense {
			t.Errorf("got %v, want %v", capturedQuery.TransactionType, valueobject.TransactionTypeExpense)
		}

		if capturedQuery.Cursor == nil || *capturedQuery.Cursor != "cursor-1" {
			t.Errorf("got %v, want %v", capturedQuery.Cursor, "cursor-1")
		}
	})
}
