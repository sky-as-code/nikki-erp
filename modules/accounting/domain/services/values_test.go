package services

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestDerefHelpersFlattenNil(t *testing.T) {
	if derefString(nil) != "" {
		t.Error("expected a nil string to flatten to empty")
	}
	if derefInt32(nil) != 0 {
		t.Error("expected a nil int32 to flatten to zero")
	}
	if derefBool(nil) {
		t.Error("expected a nil bool to flatten to false")
	}
	if !derefDecimal(nil).IsZero() {
		t.Error("expected a nil decimal to flatten to zero")
	}
	if idString(nil) != "" {
		t.Error("expected a nil id to flatten to empty")
	}
}

func TestDerefHelpersPassValuesThrough(t *testing.T) {
	text := "value"
	number := int32(7)
	flag := true
	amount := decimal.RequireFromString("12.34")

	if derefString(&text) != "value" {
		t.Error("expected the string through")
	}
	if derefInt32(&number) != 7 {
		t.Error("expected the int32 through")
	}
	if !derefBool(&flag) {
		t.Error("expected the bool through")
	}
	if !derefDecimal(&amount).Equal(amount) {
		t.Error("expected the decimal through")
	}
}

// A condition operand is stored as `any` because it compares against whatever its field holds.
// Determination compares everything as text, so every shape has to arrive as a parseable string —
// a number that came back as "1.2e+06" would fail to parse on the comparison side.
func TestConditionValueFlattensScalars(t *testing.T) {
	cases := []struct {
		raw  any
		want string
	}{
		{"standard", "standard"},
		{true, "true"},
		{false, "false"},
		{decimal.RequireFromString("8.5"), "8.5"},
		{float64(1200000), "1200000"},
		{int(42), "42"},
		{int32(42), "42"},
		{int64(42), "42"},
		{nil, ""},
	}

	for _, testCase := range cases {
		scalar, list := conditionValue(testCase.raw)
		if scalar != testCase.want {
			t.Errorf("conditionValue(%v) scalar = %q, want %q", testCase.raw, scalar, testCase.want)
		}
		if list != nil {
			t.Errorf("conditionValue(%v) unexpectedly produced a list %v", testCase.raw, list)
		}
	}
}

// A scalar operand is stored as a one-element array, because the column's jsonmap type accepts
// only an object or an array — never a bare string. That element has to come back as the scalar as
// well, or an "eq" condition reading the same column would compare against nothing.
func TestSingleElementArrayIsAlsoAScalar(t *testing.T) {
	scalar, list := conditionValue([]any{"VN_EXPORT"})

	if scalar != "VN_EXPORT" {
		t.Errorf("expected the lone element as the scalar, got %q", scalar)
	}
	if len(list) != 1 || list[0] != "VN_EXPORT" {
		t.Errorf("expected it to stay available as a list too, got %v", list)
	}
}

func TestConditionValueFlattensLists(t *testing.T) {
	scalar, list := conditionValue([]any{"a", "b", "c"})
	if scalar != "" {
		t.Errorf("expected a multi-element list to yield no scalar, got %q", scalar)
	}
	if len(list) != 3 || list[0] != "a" || list[2] != "c" {
		t.Fatalf("expected the three values, got %v", list)
	}

	if _, typed := conditionValue([]string{"x", "y"}); len(typed) != 2 {
		t.Fatalf("expected a []string through, got %v", typed)
	}
}

// A comma inside a stored JSON array element is data, not a separator — which is why a list that
// arrives already split is never re-split.
func TestConditionValueDoesNotResplitAList(t *testing.T) {
	_, list := conditionValue([]any{"Hanoi, Vietnam", "Ho Chi Minh City"})

	if len(list) != 2 {
		t.Fatalf("expected 2 values, got %d: %v", len(list), list)
	}
	if list[0] != "Hanoi, Vietnam" {
		t.Errorf("expected the comma preserved inside an element, got %q", list[0])
	}
}

func TestSplitListTrimsAndDropsEmpties(t *testing.T) {
	values := splitList("a, b,,c ")

	if len(values) != 3 {
		t.Fatalf("expected 3 values, got %d: %v", len(values), values)
	}
	if values[0] != "a" || values[1] != "b" || values[2] != "c" {
		t.Fatalf("expected trimmed values, got %v", values)
	}
	if splitList("") != nil {
		t.Error("expected an empty string to yield nil")
	}
}

func TestParseAmountTreatsGarbageAsZero(t *testing.T) {
	if !parseAmount("").IsZero() {
		t.Error("expected an empty amount to be zero")
	}
	if !parseAmount("not-a-number").IsZero() {
		t.Error("expected an unparseable amount to be zero")
	}
	if !parseAmount("12.34").Equal(decimal.RequireFromString("12.34")) {
		t.Error("expected a well-formed amount through exactly")
	}
	// Negative amounts are legitimate: a reversal is expressed as a negative charge.
	if !parseAmount("-5.5").Equal(decimal.RequireFromString("-5.5")) {
		t.Error("expected a negative amount through exactly")
	}
}
