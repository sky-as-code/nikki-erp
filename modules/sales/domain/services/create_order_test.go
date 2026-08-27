package services

import (
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The pure validation of creating an order. CreateOrder itself reads and writes the repository and
// is exercised live; the gates below are where a wrong answer would let a bad order through.

func sellableChannel() dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.SalesChannelFieldId:     "CH1",
		models.SalesChannelFieldCode:   "vdmc",
		models.SalesChannelFieldStatus: string(models.SalesChannelStatusActive),
		basemodel.FieldIsArchived:      false,
	}
}

func sellablePoint() dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.SalesPointFieldId:             "SP1",
		models.SalesPointFieldSalesChannelId: "CH1",
		models.SalesPointFieldStatus:         string(models.SalesPointStatusActive),
		basemodel.FieldIsArchived:            false,
	}
}

// Both gates must hold, and they mean different things: archived is retired for good, suspended is
// stopped for now. Either one must prevent a new sale.
func TestSellingRequiresActiveAndUnarchived(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(dmodel.DynamicFields)
		canSell bool
	}{
		{"active and unarchived", func(dmodel.DynamicFields) {}, true},
		{"archived", func(r dmodel.DynamicFields) {
			r[basemodel.FieldIsArchived] = true
		}, false},
		{"suspended", func(r dmodel.DynamicFields) {
			r[models.SalesChannelFieldStatus] = string(models.SalesChannelStatusSuspended)
		}, false},
		{"archived AND suspended", func(r dmodel.DynamicFields) {
			r[basemodel.FieldIsArchived] = true
			r[models.SalesChannelFieldStatus] = string(models.SalesChannelStatusSuspended)
		}, false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			record := sellableChannel()
			testCase.mutate(record)
			got := canSell(record, models.SalesChannelFieldStatus,
				string(models.SalesChannelStatusActive))
			if got != testCase.canSell {
				t.Errorf("canSell = %v, want %v", got, testCase.canSell)
			}
		})
	}
}

// A record with no status cannot sell. The field is required, so a missing one means a row written
// outside the sanctioned path — and defaulting it to active would let that row trade.
func TestARecordWithNoStatusCannotSell(t *testing.T) {
	if canSell(dmodel.DynamicFields{}, models.SalesChannelFieldStatus,
		string(models.SalesChannelStatusActive)) {
		t.Error("a record with no status must not be sellable")
	}
}

// A line ordering nothing is not a line (BR §55).
func TestALineMustOrderMoreThanZero(t *testing.T) {
	for _, quantity := range []string{"0", "-1"} {
		vErrs := assertLinesRequestable([]CreateOrderLine{
			{ProductVariantId: "V1", Quantity: dec(quantity)},
		})
		if vErrs == nil {
			t.Errorf("quantity %s must be refused", quantity)
			continue
		}
		if !hasReasonKey(vErrs, ReasonQuantityNotPositive) {
			t.Errorf("quantity %s must be refused as %q, got %v",
				quantity, ReasonQuantityNotPositive, vErrs.ToError())
		}
	}
}

// A line naming no variant is refused. Without inventory's product port this is the only variant
// check there is, which is why it must at least catch the empty case.
func TestALineMustNameAVariant(t *testing.T) {
	vErrs := assertLinesRequestable([]CreateOrderLine{
		{ProductVariantId: "", Quantity: dec("1")},
	})
	if vErrs == nil || !hasReasonKey(vErrs, ReasonVariantMissing) {
		t.Error("a line with no product variant must be refused")
	}
}

// Every bad line is reported, not just the first. A till submitting five lines should be told about
// all the broken ones in one round trip rather than discovering them one at a time.
func TestEveryBadLineIsReported(t *testing.T) {
	vErrs := assertLinesRequestable([]CreateOrderLine{
		{ProductVariantId: "V1", Quantity: dec("1")},
		{ProductVariantId: "", Quantity: dec("0")},
		{ProductVariantId: "V3", Quantity: dec("-2")},
	})

	if vErrs == nil {
		t.Fatal("the broken lines must be refused")
	}
	// Line 2 breaks two rules, line 3 breaks one.
	if vErrs.Count() < 3 {
		t.Errorf("expected at least 3 violations across the bad lines, got %d", vErrs.Count())
	}
}

// An order with zero lines is a valid draft (BR §69). It is confirming one that is refused.
func TestAnEmptyBasketIsAcceptable(t *testing.T) {
	if vErrs := assertLinesRequestable(nil); vErrs != nil {
		t.Errorf("an empty draft must be allowed, got %v", vErrs.ToError())
	}
}

// The violation names the line it came from, so a form can point at the offending row rather than
// saying only that something was wrong.
func TestAViolationNamesItsLine(t *testing.T) {
	vErrs := assertLinesRequestable([]CreateOrderLine{
		{ProductVariantId: "V1", Quantity: dec("1")},
		{ProductVariantId: "V2", Quantity: dec("0")},
	})

	if vErrs == nil {
		t.Fatal("the zero-quantity line must be refused")
	}
	for _, item := range *vErrs {
		if item.Field != "lines[1]" {
			t.Errorf("violation names field %q, want lines[1] — the second line is the bad one",
				item.Field)
		}
	}
}
