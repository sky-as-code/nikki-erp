package services

import (
	"testing"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The arithmetic of the initial bill, pinned without a repository. What matters here is that the
// bill allocates the WHOLE order and not a hundredth less: the balance check refuses anything else,
// and a confirm that cannot balance cannot happen at all.

func orderLineRecord(id, final, tax, quantity string) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.SalesOrderLineFieldId:              id,
		models.SalesOrderLineFieldFinalAmount:     final,
		models.SalesOrderLineFieldTaxAmount:       tax,
		models.SalesOrderLineFieldOrderedQuantity: quantity,
	}
}

// allocationsOf exercises the production allocation calculation without database writes.
func allocationsOf(
	order dmodel.DynamicFields, lines []dmodel.DynamicFields, scale int32,
) (net, tax, total map[string]decimal.Decimal) {
	allocations, _ := initialAllocations(order, lines, SalesPolicy{RoundingScale: scale})
	net = make(map[string]decimal.Decimal)
	tax = make(map[string]decimal.Decimal)
	total = make(map[string]decimal.Decimal)
	for _, entry := range allocations {
		net[entry.lineId], tax[entry.lineId], total[entry.lineId] = entry.net, entry.tax, entry.total
	}
	return net, tax, total
}

func sumOf(shares map[string]decimal.Decimal) decimal.Decimal {
	total := decimal.Zero
	for _, share := range shares {
		total = total.Add(share)
	}
	return total
}

// The bill allocates exactly the order's total. AssertOrderAllocationBalances refuses anything else
// inside the confirm transaction, so a rounding hundredth adrift here is a sale that cannot be
// confirmed at all.
func TestTheInitialBillAllocatesTheWholeOrder(t *testing.T) {
	order := dmodel.DynamicFields{
		models.SalesOrderFieldSubtotal:   "90000",
		models.SalesOrderFieldTaxTotal:   "9000",
		models.SalesOrderFieldGrandTotal: "99000",
	}
	lines := []dmodel.DynamicFields{
		orderLineRecord("L1", "33000", "3000", "1"),
		orderLineRecord("L2", "33000", "3000", "1"),
		orderLineRecord("L3", "33000", "3000", "1"),
	}

	net, tax, total := allocationsOf(order, lines, 4)

	if got := sumOf(total); !got.Equal(decimal.RequireFromString("99000")) {
		t.Errorf("allocated total = %s, want 99000", got)
	}
	if got := sumOf(net); !got.Equal(decimal.RequireFromString("90000")) {
		t.Errorf("allocated net = %s, want 90000", got)
	}
	if got := sumOf(tax); !got.Equal(decimal.RequireFromString("9000")) {
		t.Errorf("allocated tax = %s, want 9000", got)
	}
}

// Pricing rounds the grand total once, as a whole, so the line amounts can sum to slightly more or
// less than it. The allocator answers the order's number, never the lines' — copying the lines
// would leave the bill disagreeing with the sale it settles.
func TestTheBillFollowsTheOrderTotalNotTheLineAmounts(t *testing.T) {
	// Three lines of 33333.3333 sum to 99999.9999; the order was rounded to 100000.
	order := dmodel.DynamicFields{
		models.SalesOrderFieldSubtotal:   "100000",
		models.SalesOrderFieldTaxTotal:   "0",
		models.SalesOrderFieldGrandTotal: "100000",
	}
	lines := []dmodel.DynamicFields{
		orderLineRecord("L1", "33333.3333", "0", "1"),
		orderLineRecord("L2", "33333.3333", "0", "1"),
		orderLineRecord("L3", "33333.3333", "0", "1"),
	}

	_, _, total := allocationsOf(order, lines, 4)

	if got := sumOf(total); !got.Equal(decimal.RequireFromString("100000")) {
		t.Errorf("allocated total = %s, want exactly 100000: the residual must land on a line "+
			"rather than vanish", got)
	}
}

// A free line takes no money but is still on the bill: the initial bill states what was sold, and
// an item omitted from it would make the bill disagree with the order about what the customer
// bought.
func TestAFreeLineIsStillAllocatedARow(t *testing.T) {
	order := dmodel.DynamicFields{
		models.SalesOrderFieldSubtotal:   "50000",
		models.SalesOrderFieldTaxTotal:   "0",
		models.SalesOrderFieldGrandTotal: "50000",
	}
	lines := []dmodel.DynamicFields{
		orderLineRecord("PAID", "50000", "0", "1"),
		orderLineRecord("FREE", "0", "0", "1"),
	}

	_, _, total := allocationsOf(order, lines, 4)

	if _, present := total["FREE"]; !present {
		t.Error("a free line must still receive an allocation entry, of zero")
	}
	if got := total["FREE"]; !got.IsZero() {
		t.Errorf("the free line was allocated %s, want 0", got)
	}
	if got := sumOf(total); !got.Equal(decimal.RequireFromString("50000")) {
		t.Errorf("allocated total = %s, want 50000", got)
	}
}

// A sale of zero value is still a sale: it gets a bill, and that bill allocates zero rather than
// failing or being skipped.
func TestAZeroValueOrderStillProducesABalancedBill(t *testing.T) {
	order := dmodel.DynamicFields{
		models.SalesOrderFieldSubtotal:   "0",
		models.SalesOrderFieldTaxTotal:   "0",
		models.SalesOrderFieldGrandTotal: "0",
	}
	lines := []dmodel.DynamicFields{orderLineRecord("L1", "0", "0", "1")}

	_, _, total := allocationsOf(order, lines, 4)

	if got := sumOf(total); !got.IsZero() {
		t.Errorf("allocated total = %s, want 0", got)
	}
}

// The number is derived from the bill's own id, the convention create_order uses for the order
// number. Unique by construction, and a later split still derives BILL-<id>-1 from it.
func TestTheInitialBillIsNamedAfterItsOwnId(t *testing.T) {
	if got := initialBillNumberOf("01HXYZ"); got != "BILL-01HXYZ" {
		t.Errorf("bill number = %q, want BILL-01HXYZ", got)
	}

	source := dmodel.DynamicFields{
		models.SalesBillFieldBillNumber: initialBillNumberOf("01HXYZ"),
	}
	if got := billNumberOf(source, 0); got != "BILL-01HXYZ-1" {
		t.Errorf("split child number = %q, want BILL-01HXYZ-1", got)
	}
}
