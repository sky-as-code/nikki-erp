package services

import (
	"testing"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The BR §55 quantity invariant has exactly ONE enforcement point, and it is the function these
// tests cover.
//
// The plan asked for it as a database CHECK constraint as well, on the reasoning that an invariant
// this important is worth enforcing in two places. That is not available: the dynamic-model
// framework declares no CHECK constraints and no migration in either tree contains one. So there is
// no second line of defence, and these tests are correspondingly load-bearing — a gap here is a gap
// everywhere.

func lineFields(ordered, fulfilled, returned string) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.SalesOrderLineFieldOrderedQuantity:   decimal.RequireFromString(ordered),
		models.SalesOrderLineFieldFulfilledQuantity: decimal.RequireFromString(fulfilled),
		models.SalesOrderLineFieldReturnedQuantity:  decimal.RequireFromString(returned),
	}
}

func TestQuantityInvariant(t *testing.T) {
	cases := []struct {
		name      string
		ordered   string
		fulfilled string
		returned  string
		wantOk    bool
	}{
		{"a fresh line", "3", "0", "0", true},
		{"partially fulfilled", "3", "2", "0", true},
		{"fully fulfilled", "3", "3", "0", true},
		{"partially returned", "3", "3", "1", true},
		{"fully returned", "3", "3", "3", true},
		{"a fractional quantity is legitimate", "1.5", "1.5", "0.5", true},

		{"ordering nothing is not a line", "0", "0", "0", false},
		{"ordering a negative amount", "-1", "0", "0", false},
		{"fulfilling more than was ordered", "3", "4", "0", false},
		{"a negative fulfilment", "3", "-1", "0", false},
		// Measured against fulfilled rather than ordered, and the difference is the point: a
		// customer cannot return what was never handed over.
		{"returning more than was fulfilled", "3", "1", "2", false},
		{"returning from a line that fulfilled nothing", "3", "0", "1", false},
		{"a negative return", "3", "3", "-1", false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			vErrs := assertQuantitiesConsistent(
				lineFields(testCase.ordered, testCase.fulfilled, testCase.returned))
			gotOk := vErrs == nil
			if gotOk != testCase.wantOk {
				t.Errorf("ordered=%s fulfilled=%s returned=%s: consistent = %v, want %v",
					testCase.ordered, testCase.fulfilled, testCase.returned, gotOk, testCase.wantOk)
			}
		})
	}
}

// A violation must name the field the caller has to fix, not just report that something is wrong.
// A form showing "invalid line" against no particular box is not actionable.
func TestQuantityViolationNamesTheOffendingField(t *testing.T) {
	vErrs := assertQuantitiesConsistent(lineFields("3", "5", "0"))
	if vErrs == nil {
		t.Fatal("fulfilling 5 of a line that ordered 3 must be refused")
	}
	found := false
	for _, violation := range *vErrs {
		if violation.Field == models.SalesOrderLineFieldFulfilledQuantity {
			found = true
		}
	}
	if !found {
		t.Errorf("the violation must name %q, got %+v",
			models.SalesOrderLineFieldFulfilledQuantity, *vErrs)
	}
}

// The check runs against the row as it WOULD BE after the update, not against the payload alone.
//
// This is the case a payload-only check gets wrong: an update carrying just "fulfilled_quantity = 5"
// looks fine in isolation, and is a violation once merged with a stored ordered_quantity of 3.
func TestQuantityCheckUsesTheMergedRow(t *testing.T) {
	stored := lineFields("3", "0", "0")
	update := dmodel.DynamicFields{
		models.SalesOrderLineFieldFulfilledQuantity: decimal.RequireFromString("5"),
	}

	if assertQuantitiesConsistent(update) == nil {
		// Sanity: the payload alone is missing ordered_quantity, which reads as zero and so fails
		// the ordered > 0 rule. If this ever passed, the merged check below would be the only thing
		// standing between a caller and a broken row.
		t.Log("the bare payload passes on its own; the merge is what catches the real problem")
	}
	if assertQuantitiesConsistent(mergedFields(stored, update)) == nil {
		t.Error("fulfilling 5 against a stored ordered_quantity of 3 must be refused: the " +
			"invariant is a property of the resulting row, not of the payload")
	}
}

// Quantities arrive in whatever shape the caller and the round trip produced. A reader that only
// accepted decimal.Decimal would treat a JSON-supplied "3" as zero and refuse a valid line.
func TestQuantitiesAreReadFromEveryShape(t *testing.T) {
	shapes := map[string]any{
		"decimal":         decimal.RequireFromString("3"),
		"pointer decimal": func() *decimal.Decimal { d := decimal.RequireFromString("3"); return &d }(),
		"string":          "3",
		"float64":         float64(3),
		"int64":           int64(3),
		"int":             3,
	}
	for name, value := range shapes {
		t.Run(name, func(t *testing.T) {
			fields := dmodel.DynamicFields{
				models.SalesOrderLineFieldOrderedQuantity:   value,
				models.SalesOrderLineFieldFulfilledQuantity: decimal.Zero,
				models.SalesOrderLineFieldReturnedQuantity:  decimal.Zero,
			}
			if vErrs := assertQuantitiesConsistent(fields); vErrs != nil {
				t.Errorf("an ordered quantity supplied as %s must be read as 3, got %+v",
					name, *vErrs)
			}
		})
	}
}

// A malformed decimal reads as zero, which fails the ordered > 0 rule and is REPORTED as a
// violation rather than silently accepted. Silently treating it as zero would store a line that
// ordered nothing.
func TestMalformedQuantityIsRefused(t *testing.T) {
	fields := dmodel.DynamicFields{
		models.SalesOrderLineFieldOrderedQuantity:   "not a number",
		models.SalesOrderLineFieldFulfilledQuantity: decimal.Zero,
		models.SalesOrderLineFieldReturnedQuantity:  decimal.Zero,
	}
	if assertQuantitiesConsistent(fields) == nil {
		t.Error("an unparseable ordered quantity must be refused, never stored as zero")
	}
}

// The model's own view of the invariant must agree with the service's.
//
// They are two implementations of one rule — the model answers for a loaded record, the service for
// an incoming write — so a divergence would mean a line the service accepted and the model
// considered broken, or the reverse.
func TestModelAndServiceAgreeOnConsistency(t *testing.T) {
	cases := [][3]string{
		{"3", "0", "0"},
		{"3", "3", "3"},
		{"0", "0", "0"},
		{"3", "4", "0"},
		{"3", "1", "2"},
		{"1.5", "1.5", "0.5"},
	}
	for _, quantities := range cases {
		fields := lineFields(quantities[0], quantities[1], quantities[2])
		serviceSaysOk := assertQuantitiesConsistent(fields) == nil
		modelSaysOk := models.NewSalesOrderLineFrom(fields).QuantitiesAreConsistent()

		if serviceSaysOk != modelSaysOk {
			t.Errorf("ordered=%s fulfilled=%s returned=%s: service says %v, model says %v",
				quantities[0], quantities[1], quantities[2], serviceSaysOk, modelSaysOk)
		}
	}
}

// The remaining-quantity helpers never go negative. A line whose stored quantities somehow broke
// the invariant must not propagate a negative into whatever computes the next fulfilment request or
// refund — it would turn one bad row into a bad transaction.
func TestRemainingQuantitiesNeverGoNegative(t *testing.T) {
	broken := models.NewSalesOrderLineFrom(lineFields("3", "5", "9"))

	if broken.RemainingFulfillable().IsNegative() {
		t.Error("RemainingFulfillable must clamp at zero rather than report a negative")
	}
	if broken.RemainingReturnable().IsNegative() {
		t.Error("RemainingReturnable must clamp at zero rather than report a negative")
	}
}

func TestRemainingQuantities(t *testing.T) {
	line := models.NewSalesOrderLineFrom(lineFields("10", "6", "2"))

	if got := line.RemainingFulfillable(); !got.Equal(decimal.RequireFromString("4")) {
		t.Errorf("RemainingFulfillable = %s, want 4", got)
	}
	// Against fulfilled, not ordered: only what was handed over can come back.
	if got := line.RemainingReturnable(); !got.Equal(decimal.RequireFromString("4")) {
		t.Errorf("RemainingReturnable = %s, want 4", got)
	}
}

// sameFieldValue must see through the type differences a round trip introduces.
//
// A decimal read from the database and the same decimal submitted as a string are equal in value
// and different in type. A plain == would report every re-submitted snapshot as an attempt to
// change it, which would make a confirmed order unupdatable in ANY field — a read-modify-write
// cycle sends the whole line back.
func TestSameFieldValueSeesThroughTypeDifferences(t *testing.T) {
	price := decimal.RequireFromString("12.5000")

	cases := []struct {
		name      string
		stored    any
		submitted any
		want      bool
	}{
		{"identical strings", "ABC-1", "ABC-1", true},
		{"different strings", "ABC-1", "ABC-2", false},
		{"both nil", nil, nil, true},
		{"stored nil, submitted set", nil, "ABC-1", false},
		{"decimal against equal decimal", price, decimal.RequireFromString("12.5"), true},
		{"decimal against different decimal", price, decimal.RequireFromString("12.6"), false},
		{"pointer decimal against decimal", &price, decimal.RequireFromString("12.5"), true},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := sameFieldValue(testCase.stored, testCase.submitted); got != testCase.want {
				t.Errorf("sameFieldValue(%v, %v) = %v, want %v",
					testCase.stored, testCase.submitted, got, testCase.want)
			}
		})
	}
}

// The order's own editability gate: only a draft may be changed.
//
// A cancelled order is refused too, and deliberately — it records something that was attempted, and
// rewriting it would destroy the evidence of what was attempted.
func TestOnlyDraftOrdersAreEditable(t *testing.T) {
	cases := map[string]bool{
		string(models.SalesOrderStatusDraft):     true,
		string(models.SalesOrderStatusConfirmed): false,
		string(models.SalesOrderStatusCompleted): false,
		string(models.SalesOrderStatusCancelled): false,
	}
	for status, wantEditable := range cases {
		t.Run(status, func(t *testing.T) {
			order := models.NewSalesOrderFrom(dmodel.DynamicFields{
				models.SalesOrderFieldStatus: status,
			})
			if got := order.IsEditable(); got != wantEditable {
				t.Errorf("a %q order: editable = %v, want %v", status, got, wantEditable)
			}
		})
	}
}

// An archived draft is not editable either. Archiving is the system lifecycle and the status is the
// business one; a draft put away for good should not still accept line edits.
func TestArchivedDraftIsNotEditable(t *testing.T) {
	order := models.NewSalesOrderFrom(dmodel.DynamicFields{
		models.SalesOrderFieldStatus: string(models.SalesOrderStatusDraft),
		"is_archived":                true,
	})
	if order.IsEditable() {
		t.Error("an archived draft must not be editable")
	}
}

// Confirmation is what freezes the snapshots (BR §11), and a completed order counts as confirmed:
// it passed through confirmation to get there, and its receipt has been printed.
func TestIsConfirmedCoversCompleted(t *testing.T) {
	cases := map[string]bool{
		string(models.SalesOrderStatusDraft):     false,
		string(models.SalesOrderStatusConfirmed): true,
		string(models.SalesOrderStatusCompleted): true,
		string(models.SalesOrderStatusCancelled): false,
	}
	for status, wantConfirmed := range cases {
		t.Run(status, func(t *testing.T) {
			order := models.NewSalesOrderFrom(dmodel.DynamicFields{
				models.SalesOrderFieldStatus: status,
			})
			if got := order.IsConfirmed(); got != wantConfirmed {
				t.Errorf("a %q order: confirmed = %v, want %v", status, got, wantConfirmed)
			}
		})
	}
}
