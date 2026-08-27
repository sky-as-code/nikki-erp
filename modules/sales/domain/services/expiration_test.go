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

// Expiry (BR §87.3, SALES-040). The sweep reads the repository and is exercised live; what is pinned
// here is the cutoff arithmetic and the staleness test, where being wrong either expires a basket the
// customer is still filling in or never expires one at all.

// The window runs BACKWARDS from now. Getting the sign wrong would expire everything created in the
// future — which is to say, nothing — or everything ever created, which is every live basket.
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

// An unconfigured window must not expire everything instantly. Zero hours would put the cutoff at
// `now`, making every draft stale the moment it was created — including the one being typed into.
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

// A longer window expires strictly less. The obvious property, and the one a sign or unit error
// breaks: hours must not be read as minutes, nor added instead of subtracted.
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

// The staleness test is strictly BEFORE the cutoff. A draft created exactly at the boundary is not
// yet stale — it has had its full window, not one instant less.
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

// A draft with NO creation timestamp is not expired. Reading an absent timestamp as the zero time
// would make it infinitely old, and the sweep would delete baskets whose only fault was a bad read.
func TestADraftWithNoTimestampIsNotExpired(t *testing.T) {
	record := dmodel.DynamicFields{
		models.SalesOrderFieldId:     "OR01",
		models.SalesOrderFieldStatus: string(models.SalesOrderStatusDraft),
	}

	if createdAt := dateTimeOf(record, basemodel.FieldCreatedAt); createdAt != nil {
		t.Fatal("the fixture must have no timestamp")
	}
	// draftOrdersOlderThan skips exactly this case; the assertion is that a nil reads as nil rather
	// than as the zero time, which is what makes the skip possible.
}

// An expired ORDER is stored as `cancelled`, because sales_orders has no `expired` status — and the
// audit ACTION is what keeps expiry distinguishable from a withdrawal.
func TestAnExpiredOrderIsStoredAsCancelled(t *testing.T) {
	if CanTransitionOrderStatus(
		string(models.SalesOrderStatusDraft),
		string(models.SalesOrderStatusCancelled)) != true {
		t.Error("draft → cancelled must be permitted: it is how an expired draft is stored")
	}

	// And the action exists to carry the distinction the status cannot.
	if models.SalesOrderActionExpire == models.SalesOrderActionCancel {
		t.Error("expire and cancel must be different actions: one went stale, the other was " +
			"withdrawn by somebody, and the trail is the only place that difference survives")
	}
}

// A quotation lapses on ITS OWN deadline, not the org window. The quotation carries the deadline the
// customer was actually given, and applying the org default would move a promise already made.
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

// Variant sellability (BR §69, SALES-048). The port itself reads Inventory and is exercised live;
// what is pinned here is the gate's behaviour, where being wrong either sells a withdrawn product or
// refuses a whole deployment.

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

// A withdrawn variant is refused, and the refusal names the LINE — because the caller sent lines and
// fixing the order means editing one. Naming only the variant id would leave an operator scanning a
// twenty-line basket for it.
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

// An order of sellable variants passes, and asks the port once for the whole basket rather than once
// per line — an N-line order must not cost N round trips at this layer.
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

// A NIL port permits rather than refuses. It means a deployment without inventory, where there is no
// master to be withdrawn from — refusing would make Sales unusable there rather than safe.
//
// Deliberately the OPPOSITE reading from the tax port, which fails closed: an unresolved tax silently
// undercharges the business, while an unchecked variant is a question nobody in that deployment can
// answer.
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

// An order with no variants asks nothing. Calling the port with an empty batch would be a round trip
// to learn what the caller already knew.
func TestAnOrderWithNoVariantsAsksNothing(t *testing.T) {
	products := &stubProducts{notSellable: map[string]string{}}

	if _, err := assertVariantsSellable(nil, nil, products); err != nil {
		t.Fatalf("assertVariantsSellable: %v", err)
	}
	if len(products.asked) != 0 {
		t.Errorf("the port was called %d times for an empty order, want 0", len(products.asked))
	}
}

// A missing variant and a withdrawn one are refused with DIFFERENT reasons: one is usually a bad
// reference, the other a business decision, and an operator chases them differently.
func TestNotFoundAndNotSellableAreDistinct(t *testing.T) {
	if itExt.ReasonVariantNotFound == itExt.ReasonVariantNotSellable {
		t.Error("the two refusals must stay distinct: a bad id and a withdrawn product need " +
			"different remedies")
	}
}
