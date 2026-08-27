package model

import (
	"testing"

	"github.com/shopspring/decimal"
)

// GetDecimal used to bare type-assert and panic on anything that was not already a decimal.Decimal.
// A decimal crosses JSON as a STRING, so every value read back through a jsonb column crashed the
// caller. These pin the shapes that must work.

func TestGetDecimalAcceptsEveryShapeAValueCanArriveIn(t *testing.T) {
	cases := map[string]any{
		"decimal": decimal.RequireFromString("48000"),
		"string":  "48000",
		"float64": float64(48000),
		"int":     48000,
		"int64":   int64(48000),
		"int32":   int32(48000),
	}

	want := decimal.RequireFromString("48000")
	for name, value := range cases {
		t.Run(name, func(t *testing.T) {
			got := DynamicFields{"amount": value}.GetDecimal("amount")
			if got == nil {
				t.Fatalf("a %s value must be readable as a decimal", name)
			}
			if !got.Equal(want) {
				t.Errorf("%s read as %s, want %s", name, got, want)
			}
		})
	}
}

// A pointer is dereferenced rather than rejected.
func TestGetDecimalAcceptsAPointer(t *testing.T) {
	value := decimal.RequireFromString("12.34")
	got := DynamicFields{"amount": &value}.GetDecimal("amount")
	if got == nil || !got.Equal(value) {
		t.Errorf("a *decimal.Decimal must be dereferenced, got %v", got)
	}
}

// Absent, nil and unreadable all answer nil — the same answer, because a caller that handles
// "not set" handles all three correctly, and none of them is prepared to recover from a panic.
func TestGetDecimalAnswersNilRatherThanPanicking(t *testing.T) {
	cases := map[string]DynamicFields{
		"absent":      {},
		"nil":         {"amount": nil},
		"unparseable": {"amount": "not a number"},
		"wrong type":  {"amount": []string{"48000"}},
		"bool":        {"amount": true},
	}

	for name, fields := range cases {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("GetDecimal panicked on %s: %v", name, r)
				}
			}()
			if got := fields.GetDecimal("amount"); got != nil {
				t.Errorf("%s must read as nil, got %s", name, got)
			}
		})
	}
}

// Precision must survive the string path. That is the entire reason a decimal crosses JSON as a
// string rather than a number, so reading it back through float64 would defeat the point.
func TestGetDecimalKeepsPrecisionThroughTheStringPath(t *testing.T) {
	exact := "12345678901234.5678"
	got := DynamicFields{"amount": exact}.GetDecimal("amount")
	if got == nil || got.String() != exact {
		t.Errorf("precision lost: got %v, want %s", got, exact)
	}
}
