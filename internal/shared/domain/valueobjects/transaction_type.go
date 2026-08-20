package valueobjects

type TransactionType string

const (
	TransactionTypeIncome  TransactionType = "INCOME"
	TransactionTypeExpense TransactionType = "EXPENSE"
)
