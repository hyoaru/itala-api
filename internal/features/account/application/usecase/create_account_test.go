package account

import (
	"context"
	"errors"
	"testing"

	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
)

func TestCreateAccountRejectsInvalidInput(t *testing.T) {
	useCase := NewCreateAccount(newFakeRepository())

	_, err := useCase.Execute(context.Background(), CreateAccountRequest{
		UserID: "u1",
		Name:   "",
		Type:   entity.TypeAsset,
	})
	if !errors.Is(err, ErrInvalidAccount) {
		t.Fatalf("expected ErrInvalidAccount, got %v", err)
	}

	_, err = useCase.Execute(context.Background(), CreateAccountRequest{
		UserID: "u1",
		Name:   "Cash",
		Type:   entity.Type("investment"),
	})
	if !errors.Is(err, ErrInvalidAccount) {
		t.Fatalf("expected ErrInvalidAccount for invalid type, got %v", err)
	}
}

func TestCreateAccountCreatesMetadataOnly(t *testing.T) {
	useCase := NewCreateAccount(newFakeRepository())

	created, err := useCase.Execute(context.Background(), CreateAccountRequest{
		UserID: "u1",
		Name:   "Cash",
		Type:   entity.TypeAsset,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID == "" || created.Name != "Cash" || created.Type != entity.TypeAsset {
		t.Fatalf("unexpected account: %+v", created)
	}
	if created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Fatal("timestamps must be set")
	}
}
