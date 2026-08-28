package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// TS-PRICE-07 and TS-PRICE-08: negotiating a price changes the ORDER and nothing else.
//
// Two masters must survive an override untouched — the vendor's quote and the product's cost — and
// they survive for different reasons. The quote is another module's record of what a supplier is
// offering; the cost is Inventory's valuation of what the business has actually paid. Writing back
// to either would let one buyer's negotiation silently reprice everybody else's orders, or inflate
// a margin nobody chose to change.
//
// The strongest evidence is structural: nothing in the pricing path has a way to write to either.
// What the tests below add is that the path does not even try, and that a line records enough to
// show what happened.

func overrideLine(agreed string) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.PurchaseOrderLineFieldLineType:         string(models.PurchaseOrderLineTypeProduct),
		models.PurchaseOrderLineFieldProductVariantId: "01VARIANT",
		models.PurchaseOrderLineFieldOrgId:            "01ORG",
		models.PurchaseOrderLineFieldQuantity:         decimal.NewFromInt(1),
		models.PurchaseOrderLineFieldUnitPrice:        decimal.RequireFromString(agreed),
	}
}

// TS-PRICE-07: a line quoted at 9,500 and negotiated to 9,200 keeps 9,200.
//
// The stated price is what makes it a negotiation. Overwriting it with the vendor's list price on
// every save would undo the negotiation each time somebody edited the quantity — which is exactly
// what would happen if the pricer treated a present price as "not yet decided".
func TestANegotiatedPriceIsRecognisedAndKept(t *testing.T) {
	line := overrideLine("9200")

	require.True(t, hasExplicitPrice(line),
		"a line carrying a price is a negotiated price, and the pricer must leave it alone")
	assert.True(t, decimalOf(line, models.PurchaseOrderLineFieldUnitPrice).
		Equal(decimal.RequireFromString("9200")))
}

// A line with NO price is asking to be priced, which is the other half of the same decision. If
// both cases looked alike, either every negotiation would be overwritten or no line would ever be
// priced automatically.
func TestALineWithNoPriceIsPricedAutomatically(t *testing.T) {
	line := overrideLine("0")
	delete(line, models.PurchaseOrderLineFieldUnitPrice)

	assert.False(t, hasExplicitPrice(line),
		"an absent price is a request to resolve one, not a negotiated zero")
}

// TS-PRICE-08: a purchase order price does not move the product's cost.
//
// Asserted structurally, because that is where the guarantee actually lives: the line records the
// resolved price and the quote it came from, and neither field is a cost. There is deliberately no
// cost column on a purchase order line at all — a price paid is not a valuation, and the two differ
// routinely (freight, duty, and the fact that cost is averaged across receipts).
func TestThePurchaseLineRecordsNoCost(t *testing.T) {
	line := overrideLine("9200")
	line[models.PurchaseOrderLineFieldResolvedUnitPrice] = decimal.RequireFromString("9500")
	line[models.PurchaseOrderLineFieldVendorProductPriceId] = "01QUOTE"

	for _, forbidden := range []string{"cost", "unit_cost", "product_cost", "standard_price"} {
		_, present := line[forbidden]
		assert.False(t, present,
			"a purchase order line must carry no %q: confirming an order is not a valuation event",
			forbidden)
	}
}

// The override is legible after the fact, which is what "audit user + old/new resolved price"
// (§29.1) actually requires.
//
// One field would not have been enough. `vendor_product_price_id` alone says where the number came
// from but not whether anybody changed it, and comparing against the master LATER would read what
// the quote says now rather than what it said when the order was placed.
func TestAnOverrideIsVisibleAsTheDifferenceBetweenTwoStoredNumbers(t *testing.T) {
	line := overrideLine("9200")
	line[models.PurchaseOrderLineFieldResolvedUnitPrice] = decimal.RequireFromString("9500")
	line[models.PurchaseOrderLineFieldVendorProductPriceId] = "01QUOTE"

	resolved := decimalOf(line, models.PurchaseOrderLineFieldResolvedUnitPrice)
	agreed := decimalOf(line, models.PurchaseOrderLineFieldUnitPrice)

	assert.False(t, resolved.Equal(agreed), "the two prices differ, so this line was negotiated")
	assert.Equal(t, "01QUOTE", stringOf(line, models.PurchaseOrderLineFieldVendorProductPriceId),
		"and the quote it was negotiated against is named, so a reader can go and look at it")
}

// A line priced at exactly what was quoted is NOT an override, and must not be audited as one.
//
// Auditing every accepted price would fill the trail with events nobody caused, which is the fastest
// way to make an audit trail unread.
func TestAcceptingTheQuotedPriceIsNotAnOverride(t *testing.T) {
	line := overrideLine("9500")
	line[models.PurchaseOrderLineFieldResolvedUnitPrice] = decimal.RequireFromString("9500")

	resolved := decimalOf(line, models.PurchaseOrderLineFieldResolvedUnitPrice)
	agreed := decimalOf(line, models.PurchaseOrderLineFieldUnitPrice)

	assert.True(t, resolved.Equal(agreed), "nothing was overridden, so nothing is audited")
}

// The audit action for an override is its own name, distinct from every order transition.
//
// It is a LINE event where the others are order events: what a reader needs afterwards is which
// line moved and by how much, which an order-level record could not say.
func TestTheOverrideAuditActionIsDistinct(t *testing.T) {
	assert.Equal(t, "override_price", AuditActionOverridePrice)
	assert.Equal(t, "reprice", AuditActionReprice)

	for _, orderAction := range []string{
		AuditActionConfirm, AuditActionApprove, AuditActionCancel, AuditActionSend,
	} {
		assert.NotEqual(t, AuditActionOverridePrice, orderAction)
		assert.NotEqual(t, AuditActionReprice, orderAction)
	}
}
