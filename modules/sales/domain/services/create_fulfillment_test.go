package services

import (
	"testing"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// THE guard of the confirm reordering: which failed resolutions abort a confirm, and which mean
// "this is not a kiosk sale" and leave the original path alone.
//
// This is the blast radius of inverting the reserve/freeze order. Every non-kiosk sale in the system
// depends on the second answer being given for the one case that describes it.

func refusalWith(reason string) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(models.SalesOrderFulfillmentSchemaName, reason, "because"))
	return vErrs
}

// An ordinary order on a deployment with no kiosk policy resolves no method, and must confirm
// exactly as it always did. If this ever returns true, every non-kiosk confirm in the system starts
// failing.
func TestAnUnresolvedMethodOnAnUnnamedOrderDoesNotStopTheConfirm(t *testing.T) {
	if refusalStopsConfirm(false, refusalWith(ReasonMethodUnresolved)) {
		t.Fatal("an order that named no method and resolved none is an ordinary sale and must confirm")
	}
}

// An order that asked for a method by id must hear why it cannot have it — including when the answer
// is "no such method". Silently confirming it as an ordinary sale would deliver nothing and say
// nothing.
func TestANamedMethodAlwaysStopsTheConfirmOnRefusal(t *testing.T) {
	for _, reason := range []string{
		ReasonMethodUnresolved,
		ReasonMethodNotFound,
		ReasonMethodArchived,
		ReasonMethodRequiresAuth,
		ReasonTargetRequired,
		ReasonTargetNotFulfilling,
		ReasonTargetHasNoLocation,
	} {
		t.Run(reason, func(t *testing.T) {
			if !refusalStopsConfirm(true, refusalWith(reason)) {
				t.Errorf("an order naming a method must hear %q rather than confirm silently", reason)
			}
		})
	}
}

// The widening that configuration can cause. Once a channel or point carries a default method, an
// order that named nothing DOES resolve one — and a refusal about the TARGET then stops the confirm,
// because at that point it is a kiosk sale with a broken target rather than an ordinary sale.
//
// This is the behaviour to be aware of before setting a default on a live deployment: it is correct,
// and it changes which orders can confirm.
func TestATargetRefusalStopsTheConfirmEvenWhenNoMethodWasNamed(t *testing.T) {
	for _, reason := range []string{
		ReasonTargetNotFound,
		ReasonTargetNotFulfilling,
		ReasonTargetHasNoLocation,
		ReasonTargetRequired,
	} {
		t.Run(reason, func(t *testing.T) {
			if !refusalStopsConfirm(false, refusalWith(reason)) {
				t.Errorf("a resolved kiosk sale with an unusable target must refuse, not fall through")
			}
		})
	}
}

// A refusal carrying several reasons still falls through only if the survivable one is among them
// and no method was named — the unresolved reason is what identifies "not a kiosk sale", and its
// presence is the whole signal.
func TestASurvivableRefusalIsRecognisedAmongOthers(t *testing.T) {
	vErrs := refusalWith(ReasonMethodUnresolved)
	vErrs.Append(*ft.NewBusinessViolation(
		models.SalesOrderFulfillmentSchemaName, "sales_fulfillment.something_else", "and"))

	if refusalStopsConfirm(false, vErrs) {
		t.Error("the unresolved reason identifies an ordinary sale even alongside others")
	}
}

// Reuse pairs a stored fulfillment's items with the lines being confirmed. It is what makes a
// retried confirm idempotent: the fulfillment id is the reservation's source reference, so reusing
// the row reuses the hold instead of claiming the stock a second time.

func storedItem(itemId, lineId, variantId, qty string) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.SalesOrderFulfillmentItemFieldId:               itemId,
		models.SalesOrderFulfillmentItemFieldSalesOrderLineId: lineId,
		models.SalesOrderFulfillmentItemFieldProductVariantId: variantId,
		models.SalesOrderFulfillmentItemFieldOrderedQty:       decimal.RequireFromString(qty),
	}
}

func line(lineId, variantId, qty string) itExt.FulfillmentLine {
	return itExt.FulfillmentLine{
		SalesOrderLineId: lineId,
		ProductVariantId: variantId,
		Quantity:         decimal.RequireFromString(qty),
	}
}

func TestMatchingItemsAreReturnedInTheLinesOwnOrder(t *testing.T) {
	items := []dmodel.DynamicFields{
		storedItem("IT2", "L2", "V2", "1"),
		storedItem("IT1", "L1", "V1", "3"),
	}
	lines := []itExt.FulfillmentLine{line("L1", "V1", "3"), line("L2", "V2", "1")}

	matched := matchItemsToLines(items, lines)
	if matched == nil {
		t.Fatal("items that correspond to the lines must match regardless of stored order")
	}
	if matched[0] != "IT1" || matched[1] != "IT2" {
		t.Errorf("item ids must come back in the lines' order, got %v", matched)
	}
}

// THE reason quantity is part of the match. If the draft was edited between two confirm attempts,
// the old fulfillment describes a different sale — reusing it would reserve the old quantity for the
// new one, and the customer would be short.
func TestAnEditedQuantityIsNotReused(t *testing.T) {
	items := []dmodel.DynamicFields{storedItem("IT1", "L1", "V1", "3")}
	lines := []itExt.FulfillmentLine{line("L1", "V1", "5")}

	if matched := matchItemsToLines(items, lines); matched != nil {
		t.Error("a fulfillment written for a different quantity must not be reused")
	}
}

func TestAnEditedVariantIsNotReused(t *testing.T) {
	items := []dmodel.DynamicFields{storedItem("IT1", "L1", "V1", "3")}
	lines := []itExt.FulfillmentLine{line("L1", "V2", "3")}

	if matched := matchItemsToLines(items, lines); matched != nil {
		t.Error("a fulfillment written for a different product must not be reused")
	}
}

func TestADifferentLineCountIsNotReused(t *testing.T) {
	items := []dmodel.DynamicFields{storedItem("IT1", "L1", "V1", "3")}
	lines := []itExt.FulfillmentLine{line("L1", "V1", "3"), line("L2", "V2", "1")}

	if matched := matchItemsToLines(items, lines); matched != nil {
		t.Error("a fulfillment covering fewer lines than the sale must not be reused")
	}
}

// One order line may legitimately appear twice. Each stored item must be consumed at most once, or
// two identical lines would both match the same row and the second would reserve nothing.
func TestARepeatedLineConsumesEachItemOnce(t *testing.T) {
	items := []dmodel.DynamicFields{
		storedItem("IT1", "L1", "V1", "2"),
		storedItem("IT2", "L1", "V1", "2"),
	}
	lines := []itExt.FulfillmentLine{line("L1", "V1", "2"), line("L1", "V1", "2")}

	matched := matchItemsToLines(items, lines)
	if matched == nil {
		t.Fatal("two identical lines must match two identical items")
	}
	if matched[0] == matched[1] {
		t.Errorf("each stored item may be used once, got %v", matched)
	}
}
