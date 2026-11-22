package entity

type Transaction struct {
	ID                      string
	Description             string
	OccurredAt              string
	CreatedAt               string
	ReversesTransactionID   *string
	ReversedByTransactionID *string
}
