package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// TS-PRICE-04 and TS-PRICE-07/08: a confirmed transaction is not repriced by a master-data edit.
//
// This is the invariant the whole change request is built to protect. A price on a confirmed order
// is what the customer agreed to pay; a price on a purchase order is what the vendor agreed to
// charge. Editing a product afterwards must move neither, and the mechanism that guarantees it is
// the snapshot — the transaction stores its own copy rather than reading through to the master.
//
// The enforcement is split in two, and both halves are needed. The SNAPSHOT LIST says which columns
// are frozen; the status check says when. A field missing from the list is silently editable after
// confirmation, which is why the list is asserted here by name rather than by count.

// TS-PRICE-04: every price a confirmed line carries is a frozen snapshot.
//
// If `effective_unit_price` fell off this list, a confirmed order would still LOOK correct — the
// stored value would be unchanged until somebody edited the line, and then it would take a new
// price with nothing to indicate the order had been altered after the customer agreed to it.
func TestBothPricesAreSnapshotFields(t *testing.T) {
	assert.Contains(t, SnapshotFields, SalesOrderLineFieldBaseUnitPrice,
		"the base price is what the catalogue said at the time of sale")
	assert.Contains(t, SnapshotFields, SalesOrderLineFieldEffectiveUnitPrice,
		"the effective price is what the customer agreed to pay")
}

// The identity of what was sold is frozen too, not just its price.
//
// A price is only meaningful against a product: swapping the variant on a confirmed line while
// keeping the price would produce a document that charges one thing's price for another.
func TestWhatWasSoldIsFrozenAlongsideWhatItCost(t *testing.T) {
	for _, field := range []string{
		SalesOrderLineFieldProductVariantId,
		SalesOrderLineFieldProductCodeSnapshot,
		SalesOrderLineFieldProductNameSnapshot,
		SalesOrderLineFieldUomId,
		SalesOrderLineFieldTaxRateSnapshot,
	} {
		assert.Contains(t, SnapshotFields, field,
			"%s describes the moment of sale and must not change afterwards", field)
	}
}

// The list holds exactly these seven. Pinned so that ADDING a field is a deliberate act with a test
// to update, rather than something that happens by accident to a list nobody re-reads.
func TestTheSnapshotListIsExactlyTheMomentOfSale(t *testing.T) {
	require.Len(t, SnapshotFields, 7,
		"a field added here must be a decision: it becomes uneditable after confirmation")
}

// A draft order IS editable — the other half of the rule.
//
// Freezing everything would be safe and useless: a quotation exists precisely so its prices can be
// negotiated. TS-PRICE-05 depends on this being true, since repricing a draft is the supported way
// to move prices deliberately.
func TestADraftOrderIsEditable(t *testing.T) {
	order := NewSalesOrderFrom(dmodel.DynamicFields{
		SalesOrderFieldStatus: string(SalesOrderStatusDraft),
	})

	assert.True(t, order.IsEditable(), "a draft is a negotiation, not a commitment")
}

// A confirmed order is not editable, which is what refuses a reprice of one (TS-PRICE-04).
func TestAConfirmedOrderIsNotEditable(t *testing.T) {
	order := NewSalesOrderFrom(dmodel.DynamicFields{
		SalesOrderFieldStatus: string(SalesOrderStatusConfirmed),
	})

	assert.False(t, order.IsEditable(),
		"a confirmed order's prices are what the customer agreed to; repricing must be refused")
}

// An ARCHIVED draft is not editable either.
//
// Status and archival are different axes, and a check that read only the status would let an
// archived order be repriced — reviving a document somebody deliberately put away, by changing the
// numbers on it.
func TestAnArchivedDraftIsNotEditable(t *testing.T) {
	archived := true
	order := NewSalesOrderFrom(dmodel.DynamicFields{
		SalesOrderFieldStatus:     string(SalesOrderStatusDraft),
		basemodel.FieldIsArchived: archived,
	})

	assert.False(t, order.IsEditable(),
		"archival is a separate axis from status; an archived draft is still put away")
}

// An order with no status at all is not editable.
//
// Defaulting to editable would make a malformed or partially-read record the most permissive one in
// the system, which is exactly backwards.
func TestAnOrderWithNoStatusIsNotEditable(t *testing.T) {
	order := NewSalesOrderFrom(dmodel.DynamicFields{})

	assert.False(t, order.IsEditable(),
		"an unreadable status must not be read as permission")
}
