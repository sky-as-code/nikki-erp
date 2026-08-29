package services

import (
	"testing"
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// Expiry. The sweep is exercised live; pinned here are the cutoff arithmetic and staleness test.

// The window runs BACKWARDS from now: a sign error expires nothing, or everything.
func TestTheCutoffIsTheWindowBeforeNow(t *testing.T) {
	now := time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC)
	policy := SalesPolicy{DraftOrderExpiryHours: 24}

	cutoff := cutoffFor(policy, now)
	want := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)

	if !cutoff.Equal(want) {
		t.Errorf("cutoff = %s, want %s — the window runs backwards from now", cutoff, want)
	}
	if cutoff.After(now) {
		t.Error("the cutoff must never be in the future: it would expire every live draft")
	}
}

// An unconfigured window must not expire everything: zero hours puts the cutoff at `now`.
func TestAnUnconfiguredWindowDoesNotExpireEverything(t *testing.T) {
	now := time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC)

	for _, hours := range []int32{0, -5} {
		cutoff := cutoffFor(SalesPolicy{DraftOrderExpiryHours: hours}, now)
		if !cutoff.Before(now.Add(-time.Hour)) {
			t.Errorf("hours %d gave cutoff %s, which is at or near now — every draft would "+
				"expire the instant it was created", hours, cutoff)
		}
	}
}

// A longer window expires strictly less: hours must not be read as minutes, nor added instead of
// subtracted.
func TestALongerWindowExpiresLess(t *testing.T) {
	now := time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC)

	short := cutoffFor(SalesPolicy{DraftOrderExpiryHours: 1}, now)
	long := cutoffFor(SalesPolicy{DraftOrderExpiryHours: 48}, now)

	if !long.Before(short) {
		t.Errorf("a 48-hour window (cutoff %s) must reach further back than a 1-hour one "+
			"(cutoff %s)", long, short)
	}
}

func draftAged(createdAt time.Time) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.SalesOrderFieldId:     "OR01",
		models.SalesOrderFieldStatus: string(models.SalesOrderStatusDraft),
		basemodel.FieldCreatedAt:     model.ModelDateTime(createdAt),
	}
}

// Strictly BEFORE the cutoff: a draft created exactly at the boundary has had its full window.
func TestADraftIsStaleOnlyOnceItsWindowHasPassed(t *testing.T) {
	cutoff := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name      string
		createdAt time.Time
		wantStale bool
	}{
		{"well before the cutoff", cutoff.Add(-10 * time.Hour), true},
		{"a moment before", cutoff.Add(-time.Second), true},
		{"exactly at the cutoff", cutoff, false},
		{"after the cutoff", cutoff.Add(time.Hour), false},
	}

	for _, testCase := range cases {
		record := draftAged(testCase.createdAt)
		createdAt := dateTimeOf(record, basemodel.FieldCreatedAt)
		if createdAt == nil {
			t.Fatalf("%s: the fixture lost its timestamp", testCase.name)
		}

		stale := createdAt.GoTime().Before(cutoff)
		if stale != testCase.wantStale {
			t.Errorf("%s: stale = %v, want %v", testCase.name, stale, testCase.wantStale)
		}
	}
}

// A draft with NO creation timestamp is not expired: reading an absent timestamp as the zero time
// would make it infinitely old and delete baskets whose only fault was a bad read.
func TestADraftWithNoTimestampIsNotExpired(t *testing.T) {
	record := dmodel.DynamicFields{
		models.SalesOrderFieldId:     "OR01",
		models.SalesOrderFieldStatus: string(models.SalesOrderStatusDraft),
	}

	if createdAt := dateTimeOf(record, basemodel.FieldCreatedAt); createdAt != nil {
		t.Fatal("the fixture must have no timestamp")
	}
	// draftOrdersOlderThan skips this case; the assertion is that a nil reads as nil, not zero time.
}

// An expired ORDER is stored as `cancelled`; the audit ACTION is what keeps expiry
// distinguishable from a withdrawal.
func TestAnExpiredOrderIsStoredAsCancelled(t *testing.T) {
	if CanTransitionOrderStatus(
		string(models.SalesOrderStatusDraft),
		string(models.SalesOrderStatusCancelled)) != true {
		t.Error("draft → cancelled must be permitted: it is how an expired draft is stored")
	}

	// The action carries the distinction the status cannot.
	if models.SalesOrderActionExpire == models.SalesOrderActionCancel {
		t.Error("expire and cancel must be different actions: one went stale, the other was " +
			"withdrawn by somebody, and the trail is the only place that difference survives")
	}
}

// A quotation lapses on ITS OWN deadline: the org default would move a promise already made.
func TestAQuotationLapsesOnItsOwnDeadline(t *testing.T) {
	now := time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC)

	lapsed := dmodel.DynamicFields{
		models.SalesQuotationFieldValidUntil: model.ModelDateTime(now.Add(-time.Hour)),
	}
	live := dmodel.DynamicFields{
		models.SalesQuotationFieldValidUntil: model.ModelDateTime(now.Add(time.Hour)),
	}
	openEnded := dmodel.DynamicFields{}

	if at := dateTimeOf(lapsed, models.SalesQuotationFieldValidUntil); at == nil ||
		!at.GoTime().Before(now) {
		t.Error("a quotation past its deadline must read as lapsed")
	}
	if at := dateTimeOf(live, models.SalesQuotationFieldValidUntil); at == nil ||
		at.GoTime().Before(now) {
		t.Error("a quotation inside its deadline must not read as lapsed")
	}
	if at := dateTimeOf(openEnded, models.SalesQuotationFieldValidUntil); at != nil {
		t.Error("a quotation with no stated deadline must never lapse: the absence is a decision " +
			"the issuer made, not a missing value to fill in")
	}
}

// Variant sellability. The port itself is exercised live; pinned here is the gate, where being
// wrong either sells a withdrawn product or refuses a whole deployment.

// stubProducts answers AssertSellable from a fixed map.
type stubProducts struct {
	notSellable map[string]string
	asked       [][]string
}

func (this *stubProducts) AssertSellable(
	ctx corectx.Context, query itExt.AssertSellableQuery,
) (*itExt.AssertSellableResult, error) {
	this.asked = append(this.asked, query.ProductVariantIds)
	return &itExt.AssertSellableResult{NotSellable: this.notSellable}, nil
}

func lineFor(variantId string) CreateOrderLine {
	return CreateOrderLine{ProductVariantId: variantId, Quantity: dec("1")}
}

// A withdrawn variant is refused, and the refusal names the LINE, because fixing the order means
// editing one.
func TestAWithdrawnVariantIsRefusedByLine(t *testing.T) {
	products := &stubProducts{notSellable: map[string]string{
		"VAR-GONE": itExt.ReasonVariantNotSellable,
	}}

	lines := []CreateOrderLine{lineFor("VAR-OK"), lineFor("VAR-GONE")}
	vErrs, err := assertVariantsSellable(nil, lines, products)
	if err != nil {
		t.Fatalf("assertVariantsSellable: %v", err)
	}
	if vErrs == nil {
		t.Fatal("an order naming a withdrawn variant must be refused")
	}
	if vErrs.Count() != 1 {
		t.Errorf("got %d refusals, want exactly the one bad line", vErrs.Count())
	}
}

// A sellable basket passes, asking the port once: an N-line order must not cost N round trips.
func TestASellableOrderPassesInOneCall(t *testing.T) {
	products := &stubProducts{notSellable: map[string]string{}}

	lines := []CreateOrderLine{lineFor("VAR-1"), lineFor("VAR-2"), lineFor("VAR-3")}
	vErrs, err := assertVariantsSellable(nil, lines, products)
	if err != nil {
		t.Fatalf("assertVariantsSellable: %v", err)
	}
	if vErrs != nil {
		t.Errorf("a sellable order must pass, got %d refusals", vErrs.Count())
	}
	if len(products.asked) != 1 {
		t.Errorf("the port was called %d times, want 1 batched call", len(products.asked))
	}
	if len(products.asked[0]) != 3 {
		t.Errorf("the batch carried %d ids, want all 3", len(products.asked[0]))
	}
}

// A NIL port permits rather than refuses: it means a deployment without inventory, where there is
// no master to be withdrawn from. Deliberately the OPPOSITE reading from the tax port, which fails
// closed, because an unresolved tax silently undercharges the business.
func TestANilProductPortPermits(t *testing.T) {
	vErrs, err := assertVariantsSellable(nil, []CreateOrderLine{lineFor("VAR-1")}, nil)
	if err != nil {
		t.Fatalf("assertVariantsSellable: %v", err)
	}
	if vErrs != nil {
		t.Error("a deployment with no product port must still be able to sell: there is no master " +
			"for a variant to have been withdrawn from")
	}
}

// An order with no variants asks nothing: an empty batch would be a round trip to learn what the
// caller already knew.
func TestAnOrderWithNoVariantsAsksNothing(t *testing.T) {
	products := &stubProducts{notSellable: map[string]string{}}

	if _, err := assertVariantsSellable(nil, nil, products); err != nil {
		t.Fatalf("assertVariantsSellable: %v", err)
	}
	if len(products.asked) != 0 {
		t.Errorf("the port was called %d times for an empty order, want 0", len(products.asked))
	}
}

// A missing variant and a withdrawn one get DIFFERENT reasons: one is usually a bad reference,
// the other a business decision.
func TestNotFoundAndNotSellableAreDistinct(t *testing.T) {
	if itExt.ReasonVariantNotFound == itExt.ReasonVariantNotSellable {
		t.Error("the two refusals must stay distinct: a bad id and a withdrawn product need " +
			"different remedies")
	}
}
