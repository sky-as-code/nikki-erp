package services

import (
	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The bill allocation invariant: the sum of allocations across every bill of an order equals the
// order amount EXACTLY. Enforced on every bill mutation - create, split, merge - rather than only
// at settlement, because a wrong allocation found at settlement has already reached a customer.

type AllocationCheck struct {
	OrderTotal     decimal.Decimal
	AllocatedTotal decimal.Decimal

	// Difference is order minus allocated: positive means the bills do not cover the order.
	Difference decimal.Decimal

	// BillCount lets an operator tell "no bills yet" from "the bills are wrong".
	BillCount int
}

func (this AllocationCheck) Balances() bool {
	return this.Difference.IsZero()
}

const (
	ReasonAllocationMismatch = "sales_bill.allocation_mismatch"
	ReasonBillNotOpen        = "sales_bill.not_open"
	ReasonCurrencyMismatch   = "sales_bill.currency_mismatch"
	ReasonQuantityExceeded   = "sales_bill.quantity_over_allocated"
)

// CheckOrderAllocation excludes cancelled bills: they were superseded by a split or a merge and
// their allocations live on in whatever replaced them, so counting both would double every amount.
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

// AssertOrderAllocationBalances is called AFTER the write, inside the same transaction, so a
// failure rolls the whole thing back: a split produces several bills whose individual allocations
// mean nothing on their own, and only the resulting set is checkable. An order with no bills
// passes, or creating the first bill would be impossible.
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

// CheckLineAllocation is separate from the amount check because they fail differently: an amount
// mismatch is a lost residual, a quantity mismatch means the same goods went on two bills.
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
		// A cancelled bill's allocations do not count: they were superseded, not undone.
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

// AllocateAcrossBills splits one order line's value between bills in proportion to their
// quantities, so a split and the initial bill apportion the same way. The caller supplies
// AllocationInputs rather than a map so it controls the TIEBREAK: two bills taking equal
// quantities must still receive the residual reproducibly, and a map has no order.
func AllocateAcrossBills(
	lineAmount decimal.Decimal, inputs []AllocationInput, scale int32,
) map[string]decimal.Decimal {
	shares := make(map[string]decimal.Decimal, len(inputs))
	for _, result := range Allocate(lineAmount, inputs, scale) {
		shares[result.Key] = result.Amount
	}
	return shares
}

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
