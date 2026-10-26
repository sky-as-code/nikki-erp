package services

import (
	"testing"
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The pure quotation rules: the state machine and the two gates, where being wrong means honouring a
// lapsed offer or creating a second order from one acceptance. The conversion itself reads the
// repository and is exercised live.

// The transition table, written out rather than derived, so a change to the table is a change here.
func TestQuotationTransitions(t *testing.T) {
	draft := string(models.SalesQuotationStatusDraft)
	sent := string(models.SalesQuotationStatusSent)
	accepted := string(models.SalesQuotationStatusAccepted)
	expired := string(models.SalesQuotationStatusExpired)
	cancelled := string(models.SalesQuotationStatusCancelled)

	allowed := []struct{ from, to string }{
		{draft, sent},
		{draft, cancelled},
		{draft, expired},

		// Direct draft → accepted: an operator quoting on the phone may take the acceptance in the
		// same conversation.
		{draft, accepted},

		{sent, accepted},
		{sent, expired},
		{sent, cancelled},
	}
	for _, move := range allowed {
		if !CanTransitionQuotation(move.from, move.to) {
			t.Errorf("%s → %s must be allowed", move.from, move.to)
		}
	}

	forbidden := []struct {
		from, to string
		why      string
	}{
		{accepted, cancelled,
			"an accepted quotation produced an order; cancelling it would leave the order orphaned"},
		{accepted, sent,
			"a spent quotation cannot be re-offered: re-accepting would create a second order"},
		{expired, sent,
			"an expired quotation does not reopen; the customer's deadline passed and the stored " +
				"prices were never repriced"},
		{expired, accepted,
			"honouring a lapsed offer is a commercial decision that belongs in a new quotation"},
		{cancelled, sent, "a withdrawn offer is withdrawn"},
		{cancelled, accepted, "a withdrawn offer cannot be accepted"},
		{sent, draft,
			"a quotation the customer has seen cannot be un-seen; un-sending would let it be " +
				"edited while the customer holds the version they were sent"},
	}
	for _, move := range forbidden {
		if CanTransitionQuotation(move.from, move.to) {
			t.Errorf("%s → %s must be refused: %s", move.from, move.to, move.why)
		}
	}
}

// The three terminal states go nowhere; asserted as a set so a new outgoing edge on any fails here.
func TestAcceptedExpiredAndCancelledAreTerminal(t *testing.T) {
	for _, status := range []string{
		string(models.SalesQuotationStatusAccepted),
		string(models.SalesQuotationStatusExpired),
		string(models.SalesQuotationStatusCancelled),
	} {
		if next := NextQuotationStatuses(status); len(next) != 0 {
			t.Errorf("%s must be terminal, but may become %v", status, next)
		}
	}
}

// A quotation that already produced an order must report so: a second accept creating a second order
// is two deliveries and two invoices for one agreement.
func TestAConvertedQuotationIsRecognised(t *testing.T) {
	unconverted := models.NewSalesQuotationFrom(dmodel.DynamicFields{
		models.SalesQuotationFieldStatus: string(models.SalesQuotationStatusSent),
	})
	if unconverted.IsConverted() {
		t.Error("a quotation with no order recorded must not read as converted")
	}

	converted := models.NewSalesQuotationFrom(dmodel.DynamicFields{
		models.SalesQuotationFieldStatus:         string(models.SalesQuotationStatusAccepted),
		models.SalesQuotationFieldConvertedOrder: "OR01",
	})
	if !converted.IsConverted() {
		t.Error("a quotation naming an order must read as converted, so the second accept " +
			"returns that order rather than creating another")
	}
}

// An expired quotation is refused at conversion as well as by the sweep: the sweep runs on a
// schedule, so there is a window in which the stored status still says sent.
func TestConversionRefusesALapsedOffer(t *testing.T) {
	lapsed := dmodel.DynamicFields{
		models.SalesQuotationFieldStatus: string(models.SalesQuotationStatusSent),
		models.SalesQuotationFieldValidUntil: model.ModelDateTime(
			time.Now().UTC().Add(-time.Hour)),
	}

	vErrs := assertConvertible(lapsed)
	if vErrs == nil {
		t.Fatal("a quotation past valid_until must be refused even while its status says sent")
	}
	if vErrs.Count() == 0 {
		t.Error("the refusal must carry a reason an operator can act on")
	}
}

// A quotation still inside its window converts.
func TestConversionAcceptsAnOfferInsideItsWindow(t *testing.T) {
	live := dmodel.DynamicFields{
		models.SalesQuotationFieldStatus: string(models.SalesQuotationStatusSent),
		models.SalesQuotationFieldValidUntil: model.ModelDateTime(
			time.Now().UTC().Add(24 * time.Hour)),
	}

	if vErrs := assertConvertible(live); vErrs != nil {
		t.Errorf("a live quotation must convert, got %d refusals", vErrs.Count())
	}
}

// No stated expiry means no expiry rather than an expiry of zero, or every open-ended quotation would
// be refused.
func TestAQuotationWithNoStatedExpiryDoesNotLapse(t *testing.T) {
	openEnded := dmodel.DynamicFields{
		models.SalesQuotationFieldStatus: string(models.SalesQuotationStatusSent),
	}

	if vErrs := assertConvertible(openEnded); vErrs != nil {
		t.Error("a quotation with no valid_until must not be treated as expired: the absence is " +
			"a decision, not a missing value to default")
	}
}

// A cancelled quotation cannot convert, and the refusal comes from the transition table rather than a
// second hand-written rule.
func TestConversionRefusesAWithdrawnOffer(t *testing.T) {
	for _, status := range []string{
		string(models.SalesQuotationStatusCancelled),
		string(models.SalesQuotationStatusExpired),
		string(models.SalesQuotationStatusAccepted),
	} {
		record := dmodel.DynamicFields{models.SalesQuotationFieldStatus: status}
		if vErrs := assertConvertible(record); vErrs == nil {
			t.Errorf("a quotation in status %q must not convert", status)
		}
	}
}

// The conversion carries what was asked for, and the quoted price only as the engine's fallback: the
// order is repriced, not copied.
func TestConversionCarriesLinesNotTotals(t *testing.T) {
	lines := []dmodel.DynamicFields{
		{
			models.SalesQuotationLineFieldVariantId: "VAR-1",
			models.SalesQuotationLineFieldUomId:     "UOM-1",
			models.SalesQuotationLineFieldQuantity:  "3",
			models.SalesQuotationLineFieldUnitPrice: "50000",

			// Deliberately inconsistent with quantity × unit_price: if the conversion copied totals,
			// this wrong number would reach the order.
			models.SalesQuotationLineFieldFinalAmount: "999999",
		},
	}

	converted := orderLinesFromQuotation(lines)
	if len(converted) != 1 {
		t.Fatalf("converted %d lines, want 1", len(converted))
	}

	line := converted[0]
	if line.ProductVariantId != "VAR-1" || line.UomId != "UOM-1" {
		t.Error("the conversion must carry what was asked for")
	}
	if !line.Quantity.Equal(dec("3")) {
		t.Errorf("quantity = %s, want 3", line.Quantity)
	}
	if !line.UnitPrice.Equal(dec("50000")) {
		t.Errorf("unit price = %s, want 50000 as the engine's fallback", line.UnitPrice)
	}
}

// An empty quotation converts to nothing rather than to an empty order, which would be a sale of
// nothing, payable and fulfillable.
func TestAnEmptyQuotationYieldsNoLines(t *testing.T) {
	if got := orderLinesFromQuotation(nil); len(got) != 0 {
		t.Errorf("an empty quotation must yield no lines, got %d", len(got))
	}
}

// The same-status no-op is safe for send and cancel but not for accept. canTransition treats from ==
// to as allowed, which suits an idempotent status retry but not a move with a side effect: accepting
// twice would make a second order, so assertConvertible refuses it explicitly and TransitionQuotation
// refuses accepted outright.
func TestTheSameStatusNoOpIsSafeForSendAndCancelOnly(t *testing.T) {
	for _, status := range []string{
		string(models.SalesQuotationStatusSent),
		string(models.SalesQuotationStatusCancelled),
		string(models.SalesQuotationStatusExpired),
	} {
		if !CanTransitionQuotation(status, status) {
			t.Errorf("%s → %s must be an allowed no-op: a retry of a status move that already "+
				"holds has asked for a state that is true", status, status)
		}
	}

	// And the one that must not be a harmless no-op, guarded outside the table.
	accepted := dmodel.DynamicFields{
		models.SalesQuotationFieldStatus: string(models.SalesQuotationStatusAccepted),
	}
	if vErrs := assertConvertible(accepted); vErrs == nil {
		t.Error("accepting an already-accepted quotation must be refused: unlike a status retry, " +
			"it has a side effect, and a second one is a second order")
	}
}

// Manual discount gates. The repository-backed parts are exercised live; pinned here are the two
// refusals that are pure business rules.

func discountParams(amount, reason string) GrantManualDiscountParams {
	return GrantManualDiscountParams{
		SalesOrderId: "OR01",
		Amount:       dec(amount),
		Reason:       reason,
	}
}

// A reason is mandatory, enforced in the domain because it is a business invariant rather than an
// access decision: an override with no stated cause is indistinguishable from a mistake.
func TestAManualDiscountMustSayWhy(t *testing.T) {
	draft := dmodel.DynamicFields{
		models.SalesOrderFieldId:     "OR01",
		models.SalesOrderFieldStatus: string(models.SalesOrderStatusDraft),
	}

	vErrs, err := assertDiscountable(nil, draft, discountParams("1000", ""))
	if err != nil {
		t.Fatalf("assertDiscountable: %v", err)
	}
	if vErrs == nil {
		t.Fatal("an override with no reason must be refused")
	}

	// And one with a reason passes the same gate.
	vErrs, err = assertDiscountable(nil, draft, discountParams("1000", "Damaged packaging"))
	if err != nil {
		t.Fatalf("assertDiscountable: %v", err)
	}
	if vErrs != nil {
		t.Errorf("a reasoned override on a draft must be permitted, got %d refusals", vErrs.Count())
	}
}

// A negative override is a surcharge, which is not authorised: silently adding money to a customer's
// bill is refused rather than clamped.
func TestAManualDiscountCannotBeASurcharge(t *testing.T) {
	draft := dmodel.DynamicFields{
		models.SalesOrderFieldId:     "OR01",
		models.SalesOrderFieldStatus: string(models.SalesOrderStatusDraft),
	}

	for _, amount := range []string{"0", "-5000"} {
		vErrs, err := assertDiscountable(nil, draft, discountParams(amount, "Because"))
		if err != nil {
			t.Fatalf("assertDiscountable: %v", err)
		}
		if vErrs == nil {
			t.Errorf("an override of %s must be refused", amount)
		}
	}
}

// A confirmed order is frozen: discounting one would change what the customer already agreed to pay,
// possibly after a bill was raised, so the correction is a return or refund with its own money
// movement.
func TestAConfirmedOrderCannotBeDiscounted(t *testing.T) {
	for _, status := range []string{
		string(models.SalesOrderStatusConfirmed),
		string(models.SalesOrderStatusCompleted),
		string(models.SalesOrderStatusCancelled),
	} {
		record := dmodel.DynamicFields{
			models.SalesOrderFieldId:     "OR01",
			models.SalesOrderFieldStatus: status,
		}
		vErrs, err := assertDiscountable(nil, record, discountParams("1000", "Because"))
		if err != nil {
			t.Fatalf("assertDiscountable: %v", err)
		}
		if vErrs == nil {
			t.Errorf("an order in status %q is frozen and must not be discounted", status)
		}
	}
}
