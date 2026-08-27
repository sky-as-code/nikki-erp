package services

import (
	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The bill allocation invariant (BR 36, SALES-024).
//
//	Sum of allocations across every bill of an order == the order amount, EXACTLY.
//
// Exactly, not approximately. The word matters: allocations are produced by dividing an order across
// bills, and dividing rarely comes out even. Rounding each share independently and hoping is how a
// business ends up with a customer who paid three bills that sum to one dong less than the sale.
// The D-04 allocator assigns the whole residual rather than letting it vanish, and this file is what
// checks the result.
//
// It is enforced on every bill mutation - create, split, merge - rather than only at settlement,
// because a wrong allocation discovered at settlement has already been shown to a customer.

// AllocationCheck is what a verification concluded.
type AllocationCheck struct {
	OrderTotal     decimal.Decimal
	AllocatedTotal decimal.Decimal

	// Difference is order minus allocated: positive means the bills do not cover the order, negative
	// means they cover more than it. Carried rather than recomputed so a caller reporting the
	// failure names the same number the check used.
	Difference decimal.Decimal

	// BillCount is how many live bills were summed, so an operator can tell "no bills yet" from
	// "the bills are wrong".
	BillCount int
}

// Balances reports whether the invariant holds.
func (this AllocationCheck) Balances() bool {
	return this.Difference.IsZero()
}

// The refusal reasons the allocation rules produce.
const (
	ReasonAllocationMismatch = "sales_bill.allocation_mismatch"
	ReasonBillNotOpen        = "sales_bill.not_open"
	ReasonCurrencyMismatch   = "sales_bill.currency_mismatch"
	ReasonQuantityExceeded   = "sales_bill.quantity_over_allocated"
)

// CheckOrderAllocation verifies BR 36 for one order.
//
// Cancelled bills are excluded. They were superseded by a split or a merge and their allocations
// live on in whatever replaced them; counting both would double every amount the operation touched
// and report a failure that is really an artefact of keeping history (BR 83).
func CheckOrderAllocation(
	ctx corectx.Context, orderId string,
) (*AllocationCheck, error) {
	orderRecord, err := loadRecord(ctx,
		models.SalesOrderSchemaName, models.SalesOrderFieldId, orderId)
	if err != nil || orderRecord == nil {
		return nil, err
	}

	bills, err := searchBy(ctx,
		models.SalesBillSchemaName, models.SalesBillFieldSalesOrderId, orderId)
	if err != nil {
		return nil, err
	}

	allocated := decimal.Zero
	live := 0
	for _, bill := range bills {
		if stringOf(bill, models.SalesBillFieldStatus) ==
			string(models.SalesBillStatusCancelled) {
			continue
		}
		live++

		billId := stringOf(bill, models.SalesBillFieldId)
		allocations, err := searchBy(ctx,
			models.SalesBillLineSchemaName, models.SalesBillLineFieldSalesBillId, billId)
		if err != nil {
			return nil, err
		}
		allocated = allocated.Add(models.SumAllocatedTotal(allocations))
	}

	orderTotal := decimalOf(orderRecord, models.SalesOrderFieldGrandTotal)
	return &AllocationCheck{
		OrderTotal:     orderTotal,
		AllocatedTotal: allocated,
		Difference:     orderTotal.Sub(allocated),
		BillCount:      live,
	}, nil
}

// AssertOrderAllocationBalances refuses a mutation that would break BR 36.
//
// Called AFTER the write, inside the same transaction, so a failure rolls the whole thing back. That
// is the opposite of the usual validate-then-write shape and is deliberate: a split produces several
// bills whose individual allocations mean nothing on their own, and the only checkable statement is
// about the set they form once written.
//
// An order with no bills at all passes. A sale that has not been billed yet has nothing to reconcile,
// and refusing it would make creating the first bill impossible.
func AssertOrderAllocationBalances(
	ctx corectx.Context, orderId string,
) (*ft.ClientErrors, error) {
	check, err := CheckOrderAllocation(ctx, orderId)
	if err != nil || check == nil {
		return nil, err
	}
	if check.BillCount == 0 || check.Balances() {
		return nil, nil
	}

	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation("allocations", ReasonAllocationMismatch,
		"the bills of this order allocate "+check.AllocatedTotal.String()+
			" against an order total of "+check.OrderTotal.String()+
			"; every dong of a sale must be settled by exactly one bill"))
	return vErrs, nil
}

// CheckLineAllocation verifies that one order line is not over-allocated across bills.
//
// Separate from the amount check because they fail differently and an operator fixes them
// differently: an amount mismatch is a rounding or a lost residual, a quantity mismatch means the
// same goods were put on two bills. BR 36 states the amount rule; BR 37 forbids the double
// allocation a split could otherwise produce.
func CheckLineAllocation(
	ctx corectx.Context, orderLineId string,
) (ordered decimal.Decimal, allocated decimal.Decimal, err error) {
	lineRecord, err := loadRecord(ctx,
		models.SalesOrderLineSchemaName, models.SalesOrderLineFieldId, orderLineId)
	if err != nil || lineRecord == nil {
		return decimal.Zero, decimal.Zero, err
	}
	ordered = decimalOf(lineRecord, models.SalesOrderLineFieldOrderedQuantity)

	allocations, err := searchBy(ctx,
		models.SalesBillLineSchemaName, models.SalesBillLineFieldSalesOrderLineId, orderLineId)
	if err != nil {
		return ordered, decimal.Zero, err
	}

	allocated = decimal.Zero
	for _, allocation := range allocations {
		// A cancelled bill's allocations do not count, for the same reason they do not in the amount
		// check: they were superseded, not undone.
		billId := stringOf(allocation, models.SalesBillLineFieldSalesBillId)
		bill, err := loadRecord(ctx, models.SalesBillSchemaName, models.SalesBillFieldId, billId)
		if err != nil {
			return ordered, allocated, err
		}
		if bill == nil || stringOf(bill, models.SalesBillFieldStatus) ==
			string(models.SalesBillStatusCancelled) {
			continue
		}
		allocated = allocated.Add(decimalOf(allocation, models.SalesBillLineFieldQuantity))
	}
	return ordered, allocated, nil
}

// AssertLineNotOverAllocated refuses an allocation that would put more of a line on bills than the
// order actually contains.
func AssertLineNotOverAllocated(
	ctx corectx.Context, orderLineId string,
) (*ft.ClientErrors, error) {
	ordered, allocated, err := CheckLineAllocation(ctx, orderLineId)
	if err != nil {
		return nil, err
	}
	if !allocated.GreaterThan(ordered) {
		return nil, nil
	}

	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation("quantity", ReasonQuantityExceeded,
		"this order line has "+allocated.String()+" allocated across bills but only "+
			ordered.String()+" was ordered; the same goods cannot be billed twice"))
	return vErrs, nil
}

// AllocateAcrossBills splits one order line's value between bills in proportion to their quantities.
//
// It exists so that split (SALES-025) and the initial bill both apportion the same way. The D-04
// residual rule guarantees the shares sum to the line's amount EXACTLY, which is what makes BR 36
// checkable rather than merely approximately true.
//
// The caller supplies AllocationInputs rather than a map so it controls the TIEBREAK: two bills
// taking equal quantities must still receive the residual reproducibly, and a map has no order to
// break the tie with.
func AllocateAcrossBills(
	lineAmount decimal.Decimal, inputs []AllocationInput, scale int32,
) map[string]decimal.Decimal {
	shares := make(map[string]decimal.Decimal, len(inputs))
	for _, result := range Allocate(lineAmount, inputs, scale) {
		shares[result.Key] = result.Amount
	}
	return shares
}

// billsOfOrder reads an order's live bills, newest constraint first.
func billsOfOrder(
	ctx corectx.Context, orderId string, includeCancelled bool,
) ([]dmodel.DynamicFields, error) {
	bills, err := searchBy(ctx,
		models.SalesBillSchemaName, models.SalesBillFieldSalesOrderId, orderId)
	if err != nil {
		return nil, err
	}
	if includeCancelled {
		return bills, nil
	}

	live := make([]dmodel.DynamicFields, 0, len(bills))
	for _, bill := range bills {
		if stringOf(bill, models.SalesBillFieldStatus) !=
			string(models.SalesBillStatusCancelled) {
			live = append(live, bill)
		}
	}
	return live, nil
}
