package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// A confirmed transaction is not repriced by a master-data edit: the transaction stores its own
// snapshot rather than reading through to the master. Enforcement has two halves — the snapshot list
// says which columns are frozen, the status check says when — so the list is asserted by name rather
// than by count, since a field missing from it is silently editable after confirmation.

// Every price a confirmed line carries is a frozen snapshot. A field falling off this list would
// still look correct until somebody edited the line, and then it would take a new price with nothing
// to indicate the order had been altered after the customer agreed to it.
func TestBothPricesAreSnapshotFields(t *testing.T) {
	assert.Contains(t, SnapshotFields, SalesOrderLineFieldBaseUnitPrice,
		"the base price is what the catalogue said at the time of sale")
	assert.Contains(t, SnapshotFields, SalesOrderLineFieldEffectiveUnitPrice,
		"the effective price is what the customer agreed to pay")
}

// The identity of what was sold is frozen too: swapping the variant on a confirmed line while
// keeping the price would charge one thing's price for another.
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

// The list holds exactly these seven, pinned so adding a field is a deliberate act with a test to
// update.
func TestTheSnapshotListIsExactlyTheMomentOfSale(t *testing.T) {
	require.Len(t, SnapshotFields, 7,
		"a field added here must be a decision: it becomes uneditable after confirmation")
}

// A draft order is editable — the other half of the rule. Repricing a draft is the supported way to
// move prices deliberately.
func TestADraftOrderIsEditable(t *testing.T) {
	order := NewSalesOrderFrom(dmodel.DynamicFields{
		SalesOrderFieldStatus: string(SalesOrderStatusDraft),
	})

	assert.True(t, order.IsEditable(), "a draft is a negotiation, not a commitment")
}

// A confirmed order is not editable, which is what refuses a reprice of one.
func TestAConfirmedOrderIsNotEditable(t *testing.T) {
	order := NewSalesOrderFrom(dmodel.DynamicFields{
		SalesOrderFieldStatus: string(SalesOrderStatusConfirmed),
	})

	assert.False(t, order.IsEditable(),
		"a confirmed order's prices are what the customer agreed to; repricing must be refused")
}

// An archived draft is not editable either. Status and archival are different axes, and a check
// reading only the status would let an archived order be repriced.
func TestAnArchivedDraftIsNotEditable(t *testing.T) {
	archived := true
	order := NewSalesOrderFrom(dmodel.DynamicFields{
		SalesOrderFieldStatus:     string(SalesOrderStatusDraft),
		basemodel.FieldIsArchived: archived,
	})

	assert.False(t, order.IsEditable(),
		"archival is a separate axis from status; an archived draft is still put away")
}

// An order with no status at all is not editable: defaulting to editable would make a malformed
// record the most permissive one in the system.
func TestAnOrderWithNoStatusIsNotEditable(t *testing.T) {
	order := NewSalesOrderFrom(dmodel.DynamicFields{})

	assert.False(t, order.IsEditable(),
		"an unreadable status must not be read as permission")
}
