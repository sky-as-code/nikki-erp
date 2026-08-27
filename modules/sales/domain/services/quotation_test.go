package services

import (
	"testing"
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The quotation rules that are pure. The conversion itself reads the repository and is exercised
// live; what is pinned here is the state machine and the two gates, where being wrong means either
// honouring an offer that lapsed or creating a second order from one acceptance.

// The transition table, in full. Written out rather than derived, so that a change to the table is a
// change to this test — which is the point of having the table at all.
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

		// Direct draft → accepted: a back-office operator quoting on the phone may take the
		// acceptance in the same conversation, and forcing a send in between would record a step
		// that did not happen.
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

// The three terminal states go nowhere. Asserted as a set rather than one by one, so a new outgoing
// edge added to any of them fails here.
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

// THE idempotency check of the conversion. A quotation that already produced an order must report
// so, because a second accept creating a second order is two deliveries and two invoices for one
// agreement.
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

// An expired quotation is refused at conversion, IN ADDITION to being caught by the sweep. The sweep
// runs on a schedule, so between a quotation lapsing and the sweep noticing there is a window in
// which the stored status still says `sent` — and converting inside it would honour an offer that
// had already expired.
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

// No stated expiry means no expiry, rather than an expiry of zero. A null valid_until is a decision
// the issuer made, and reading it as "already lapsed" would refuse every open-ended quotation.
func TestAQuotationWithNoStatedExpiryDoesNotLapse(t *testing.T) {
	openEnded := dmodel.DynamicFields{
		models.SalesQuotationFieldStatus: string(models.SalesQuotationStatusSent),
	}

	if vErrs := assertConvertible(openEnded); vErrs != nil {
		t.Error("a quotation with no valid_until must not be treated as expired: the absence is " +
			"a decision, not a missing value to default")
	}
}

// A cancelled quotation cannot convert, and the refusal comes from the transition table rather than
// from a second hand-written rule — so the table stays the single source of truth.
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

// The conversion carries WHAT was asked for, and the quoted price only as the engine's fallback.
// This is the file's central decision: the order is repriced, not copied.
func TestConversionCarriesLinesNotTotals(t *testing.T) {
	lines := []dmodel.DynamicFields{
		{
			models.SalesQuotationLineFieldVariantId: "VAR-1",
			models.SalesQuotationLineFieldUomId:     "UOM-1",
			models.SalesQuotationLineFieldQuantity:  "3",
			models.SalesQuotationLineFieldUnitPrice: "50000",

			// Deliberately inconsistent with quantity × unit_price. If the conversion copied
			// totals, this wrong number would reach the order; because it reprices, it cannot.
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

// An empty quotation converts to nothing, and the conversion must say so rather than creating an
// empty order — which would be a sale of nothing, payable and fulfillable.
func TestAnEmptyQuotationYieldsNoLines(t *testing.T) {
	if got := orderLinesFromQuotation(nil); len(got) != 0 {
		t.Errorf("an empty quotation must yield no lines, got %d", len(got))
	}
}

// The same-status no-op is safe for `send` and `cancel` and would NOT be for `accept`.
//
// canTransition treats from == to as allowed, which is right for an idempotent retry of a status
// move and wrong for anything with a side effect. Re-sending a sent quotation shows the customer the
// same document; re-cancelling a cancelled one changes nothing. Accepting twice would make a second
// order — which is why assertConvertible refuses it explicitly rather than relying on the table, and
// why TransitionQuotation refuses `accepted` outright.
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

	// And the one that must not be treated as a harmless no-op, guarded outside the table.
	accepted := dmodel.DynamicFields{
		models.SalesQuotationFieldStatus: string(models.SalesQuotationStatusAccepted),
	}
	if vErrs := assertConvertible(accepted); vErrs == nil {
		t.Error("accepting an already-accepted quotation must be refused: unlike a status retry, " +
			"it has a side effect, and a second one is a second order")
	}
}

// Manual discount gates (BR §87.4, SALES-039). The repository-backed parts are exercised live; what
// is pinned here is the two refusals that are pure business rules.

func discountParams(amount, reason string) GrantManualDiscountParams {
	return GrantManualDiscountParams{
		SalesOrderId: "OR01",
		Amount:       dec(amount),
		Reason:       reason,
	}
}

// A reason is MANDATORY. It is enforced in the domain rather than in app/ because it is a business
// invariant, not an access decision — an override with no stated cause is indistinguishable from a
// mistake, and it is the field an auditor asking why this customer paid less actually reads.
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

// A negative override is a SURCHARGE, which BR §87.4 does not authorise. Silently adding money to a
// customer's bill is the worst failure available here, so it is refused rather than clamped.
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

// A CONFIRMED order is frozen (BR §11). Discounting one would change what a customer already agreed
// to pay, after a bill may already have been raised — so the correction after confirmation is a
// return or a refund, each with its own money movement, not a retrospective price edit.
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
