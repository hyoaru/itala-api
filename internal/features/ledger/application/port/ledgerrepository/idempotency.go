package ledgerrepository

type IdempotencyRecord struct {
	TransactionID string
	RequestHash   string
	CreatedAt     string
}
