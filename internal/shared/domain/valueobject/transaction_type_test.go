package valueobject

import "testing"

func TestTransactionType_IsValid(t *testing.T) {
	t.Run("income", func(t *testing.T) {
		if !TransactionTypeIncome.IsValid() {
			t.Error("got false, want true")
		}
	})

	t.Run("expense", func(t *testing.T) {
		if !TransactionTypeExpense.IsValid() {
			t.Error("got false, want true")
		}
	})

	t.Run("invalid", func(t *testing.T) {
		if TransactionType("TRANSFER").IsValid() {
			t.Error("got true, want false")
		}
	})

	t.Run("empty", func(t *testing.T) {
		if TransactionType("").IsValid() {
			t.Error("got true, want false")
		}
	})
}
