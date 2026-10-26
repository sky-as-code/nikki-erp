package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// Negotiating a price changes the order and nothing else. The vendor's quote and the product's cost
// must both survive untouched: writing back to either would let one buyer's negotiation reprice
// everybody else's orders or move a valuation nobody chose to change.

func overrideLine(agreed string) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.PurchaseOrderLineFieldLineType:         string(models.PurchaseOrderLineTypeProduct),
		models.PurchaseOrderLineFieldProductVariantId: "01VARIANT",
		models.PurchaseOrderLineFieldOrgId:            "01ORG",
		models.PurchaseOrderLineFieldQuantity:         decimal.NewFromInt(1),
		models.PurchaseOrderLineFieldUnitPrice:        decimal.RequireFromString(agreed),
	}
}

// A line quoted at 9,500 and negotiated to 9,200 keeps 9,200: a stated price is a negotiation, and
// treating it as "not yet decided" would undo it on every save.
func TestANegotiatedPriceIsRecognisedAndKept(t *testing.T) {
	line := overrideLine("9200")

	require.True(t, hasExplicitPrice(line),
		"a line carrying a price is a negotiated price, and the pricer must leave it alone")
	assert.True(t, decimalOf(line, models.PurchaseOrderLineFieldUnitPrice).
		Equal(decimal.RequireFromString("9200")))
}

// A line with no price is asking to be priced — the other half of the same decision.
func TestALineWithNoPriceIsPricedAutomatically(t *testing.T) {
	line := overrideLine("0")
	delete(line, models.PurchaseOrderLineFieldUnitPrice)

	assert.False(t, hasExplicitPrice(line),
		"an absent price is a request to resolve one, not a negotiated zero")
}

// A purchase order price does not move the product's cost. Asserted structurally: the line records
// the resolved price and the quote it came from, and there is deliberately no cost column at all —
// a price paid is not a valuation, given freight, duty and averaging across receipts.
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

// The override stays legible after the fact. One field would not do: vendor_product_price_id alone
// says where the number came from but not whether anybody changed it, and comparing against the
// master later reads what the quote says now, not what it said when the order was placed.
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

// A line priced at exactly what was quoted is not an override, and auditing it would fill the trail
// with events nobody caused.
func TestAcceptingTheQuotedPriceIsNotAnOverride(t *testing.T) {
	line := overrideLine("9500")
	line[models.PurchaseOrderLineFieldResolvedUnitPrice] = decimal.RequireFromString("9500")

	resolved := decimalOf(line, models.PurchaseOrderLineFieldResolvedUnitPrice)
	agreed := decimalOf(line, models.PurchaseOrderLineFieldUnitPrice)

	assert.True(t, resolved.Equal(agreed), "nothing was overridden, so nothing is audited")
}

// The audit action for an override is its own name: it is a line event where the others are order
// events, and a reader needs to know which line moved and by how much.
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
