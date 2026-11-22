package account

import (
	"strings"
	"time"

	"golang.org/x/text/unicode/norm"
)

func NormalizeName(name string) string {
	return strings.ToLower(norm.NFKC.String(strings.TrimSpace(name)))
}

type Type string

const (
	TypeAsset     Type = "asset"
	TypeLiability Type = "liability"
	TypeIncome    Type = "income"
	TypeExpense   Type = "expense"
)

func (t Type) IsValid() bool {
	switch t {
	case TypeAsset, TypeLiability, TypeIncome, TypeExpense:
		return true
	default:
		return false
	}
}

type Account struct {
	ID        string
	Name      string
	Type      Type
	DeletedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
