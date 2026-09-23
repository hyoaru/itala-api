package valueobject

import (
	"errors"
	"testing"
)

func TestNewDecimal(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		got, err := NewDecimal("10.5")
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if got.String() != "10.5" {
			t.Errorf("got %v, want %v", got.String(), "10.5")
		}
	})

	t.Run("invalid", func(t *testing.T) {
		if _, err := NewDecimal("abc"); err == nil {
			t.Errorf("got %v, want non-nil", err)
		}
	})
}

func TestDecimal_Add(t *testing.T) {
	left, _ := NewDecimal("10")
	right, _ := NewDecimal("20")

	if got := left.Add(right).String(); got != "30" {
		t.Errorf("got %v, want %v", got, "30")
	}
}

func TestDecimal_Subtract(t *testing.T) {
	left, _ := NewDecimal("10")
	right, _ := NewDecimal("20")

	if got := left.Subtract(right).String(); got != "-10" {
		t.Errorf("got %v, want %v", got, "-10")
	}
}

func TestDecimal_Negate(t *testing.T) {
	value, _ := NewDecimal("10")

	if got := value.Negate().String(); got != "-10" {
		t.Errorf("got %v, want %v", got, "-10")
	}
}

func TestDecimal_Multiply(t *testing.T) {
	left, _ := NewDecimal("10")
	right, _ := NewDecimal("20")

	if got := left.Multiply(right).String(); got != "200" {
		t.Errorf("got %v, want %v", got, "200")
	}
}

func TestDecimal_Divide(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		left, _ := NewDecimal("20")
		right, _ := NewDecimal("10")

		got, err := left.Divide(right)
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if got.String() != "2" {
			t.Errorf("got %v, want %v", got.String(), "2")
		}
	})

	t.Run("division by zero", func(t *testing.T) {
		left, _ := NewDecimal("10")
		right, _ := NewDecimal("0")

		_, err := left.Divide(right)

		if !errors.Is(err, ErrDivisionByZero) {
			t.Errorf("got %v, want %v", err, ErrDivisionByZero)
		}
	})
}

func TestDecimal_UnmarshalJSON(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		var got Decimal

		if err := got.UnmarshalJSON([]byte("12.34")); err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if got.String() != "12.34" {
			t.Errorf("got %v, want %v", got.String(), "12.34")
		}
	})

	t.Run("invalid", func(t *testing.T) {
		var got Decimal

		if err := got.UnmarshalJSON([]byte("abc")); err == nil {
			t.Errorf("got %v, want non-nil", err)
		}
	})
}
