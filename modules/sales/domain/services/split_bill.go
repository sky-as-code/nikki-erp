package services

import (
	"time"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	lock "github.com/sky-as-code/nikki-erp/modules/core/infra/distributedlock"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// Splitting a bill: one settlement unit becomes several.
//
// Total before equals total after. The allocator apportions each line with the residual rule so the
// parts sum to the whole exactly, and the balance check runs afterwards inside the same transaction
// so a split that failed to balance rolls back.
//
// A split is NOT a pricing event: promotions are not recalculated, prices are not re-resolved and
// the order is not repriced. Splitting can change which promotions WOULD apply if the basket were
// re-evaluated, and re-evaluating would silently rewrite what the customer already agreed to.
//
// The source bill is never deleted. It is marked cancelled and lineage rows point from it to each
// bill it became, because a payment already recorded against it must still resolve to something.

// SplitBillParams describes how to divide a bill.
type SplitBillParams struct {
	SourceBillId string

	// Parts are the bills to create. Each names how much of each order line it takes.
	Parts []SplitBillPart
}

// SplitBillPart is one resulting bill.
type SplitBillPart struct {
	// Allocations maps sales_order_line_id to the quantity this part settles.
	Allocations map[string]decimal.Decimal
}

// SplitBillResult is what a split produced.
type SplitBillResult struct {
	SourceBillId string

	// CreatedBillIds are the new bills, in the order the parts were given.
	CreatedBillIds []string

	// TotalBefore and TotalAfter are carried so a caller can see the invariant held rather than
	// taking it on trust.
	TotalBefore decimal.Decimal
	TotalAfter  decimal.Decimal
}

// The refusal reasons split can produce.
const (
	ReasonBillNotFound         = "sales_bill.not_found"
	ReasonSplitNeedsTwoParts   = "sales_bill.split_needs_two_parts"
	ReasonAllocationIncomplete = "sales_bill.allocation_incomplete"
	ReasonSplitTotalChanged    = "sales_bill.split_total_changed"
)

// SplitBill divides one open bill into several.
//
// Under the ORDER's lock rather than the bill's: a split touches every bill of the order through
// the balance invariant, and locking one bill would let a concurrent split of a sibling produce a
// set that individually balances and collectively does not.
func SplitBill(
	ctx corectx.Context, params SplitBillParams, dLock lock.DistributedLock, policy SalesPolicy,
) (*SplitBillResult, *ft.ClientErrors, error) {
	if dLock == nil {
		return nil, nil, errors.New(
			"the distributed lock is not available; a bill cannot be split without it")
	}

	source, err := loadRecord(ctx,
		models.SalesBillSchemaName, models.SalesBillFieldId, params.SourceBillId)
	if err != nil {
		return nil, nil, err
	}
	if source == nil {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("source_bill_id", ReasonBillNotFound,
			"no bill exists with id '"+params.SourceBillId+"'"))
		return nil, vErrs, nil
	}

	orderId := stringOf(source, models.SalesBillFieldSalesOrderId)
	key := confirmLockKeyOf(orderId)
	acquired, err := dLock.AcquireWithRetry(
		ctx, key, confirmLockTtl, confirmLockRetryCount, confirmLockRetryDelay)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "acquiring the lock of order '%s'", orderId)
	}
	if !acquired {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("source_bill_id", ReasonLockUnavailable,
			"this order is being changed by another request; try again"))
		return nil, vErrs, nil
	}
	defer func() { _ = dLock.Release(ctx, key) }()

	// Re-read under the lock: the record fetched above described the world as it was while we were
	// still queuing, and its status may have moved.
	source, err = loadRecord(ctx,
		models.SalesBillSchemaName, models.SalesBillFieldId, params.SourceBillId)
	if err != nil || source == nil {
		return nil, nil, err
	}

	return splitUnderLock(ctx, source, params, policy)
}

func splitUnderLock(
	ctx corectx.Context,
	source dmodel.DynamicFields,
	params SplitBillParams,
	policy SalesPolicy,
) (*SplitBillResult, *ft.ClientErrors, error) {
	if vErrs := assertSplittable(source, params); vErrs != nil {
		return nil, vErrs, nil
	}

	orderId := stringOf(source, models.SalesBillFieldSalesOrderId)
	sourceId := stringOf(source, models.SalesBillFieldId)

	existing, err := searchBy(ctx,
		models.SalesBillLineSchemaName, models.SalesBillLineFieldSalesBillId, sourceId)
	if err != nil {
		return nil, nil, err
	}
	if vErrs := assertAllocationsCoverSource(existing, params); vErrs != nil {
		return nil, vErrs, nil
	}

	totalBefore := decimalOf(source, models.SalesBillFieldTotalAmount)

	var created []string
	var totalAfter decimal.Decimal
	err = withTransaction(ctx, models.SalesBillSchemaName, func(tranxCtx corectx.Context) error {
		created, totalAfter, err = writeSplitParts(tranxCtx, source, existing, params, policy)
		if err != nil {
			return err
		}
		if err := cancelSupersededBill(tranxCtx, source, created,
			string(models.SalesBillRelationSplitInto)); err != nil {
			return err
		}

		// Checked inside the transaction so a split that does not balance rolls back whole.
		vErrs, err := AssertOrderAllocationBalances(tranxCtx, orderId)
		if err != nil {
			return err
		}
		if vErrs != nil {
			return errors.New("the split left the order's bills unbalanced: " +
				vErrs.ToError().Error())
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	// Total before equals total after. Asserted rather than assumed: the allocator guarantees each
	// LINE apportions exactly, and this is the statement about the bills as a set.
	if !totalAfter.Equal(totalBefore) {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("parts", ReasonSplitTotalChanged,
			"the split produced "+totalAfter.String()+" from a bill of "+totalBefore.String()))
		return nil, vErrs, nil
	}

	return &SplitBillResult{
		SourceBillId:   sourceId,
		CreatedBillIds: created,
		TotalBefore:    totalBefore,
		TotalAfter:     totalAfter,
	}, nil, nil
}

// assertSplittable applies the gates that come before a split.
func assertSplittable(source dmodel.DynamicFields, params SplitBillParams) *ft.ClientErrors {
	refuse := func(field, reason, message string) *ft.ClientErrors {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation(field, reason, message))
		return vErrs
	}

	if !models.NewSalesBillFrom(source).IsOpen() {
		// A settled bill has money against it and a cancelled one is already superseded. Splitting
		// either would leave a payment pointing at a bill that no longer represents what was paid.
		return refuse("source_bill_id", ReasonBillNotOpen,
			"only an open bill may be split; this one is '"+
				stringOf(source, models.SalesBillFieldStatus)+"'")
	}

	if len(params.Parts) < 2 {
		// Splitting into one is not a split, and splitting into none would destroy the bill. Both are
		// refused rather than treated as a no-op, because either is a caller mistake.
		return refuse("parts", ReasonSplitNeedsTwoParts,
			"a split must produce at least two bills")
	}
	return nil
}

// assertAllocationsCoverSource checks the split accounts for exactly what the source bill held.
// Both directions matter: under-allocating leaves value on no bill at all so the customer is never
// asked for it, and over-allocating bills the same goods twice. Neither shows up in one part.
func assertAllocationsCoverSource(
	existing []dmodel.DynamicFields, params SplitBillParams,
) *ft.ClientErrors {
	wanted := make(map[string]decimal.Decimal, len(existing))
	for _, allocation := range existing {
		lineId := stringOf(allocation, models.SalesBillLineFieldSalesOrderLineId)
		wanted[lineId] = wanted[lineId].Add(
			decimalOf(allocation, models.SalesBillLineFieldQuantity))
	}

	given := make(map[string]decimal.Decimal, len(wanted))
	for _, part := range params.Parts {
		for lineId, quantity := range part.Allocations {
			given[lineId] = given[lineId].Add(quantity)
		}
	}

	vErrs := ft.NewClientErrors()
	for lineId, want := range wanted {
		got := given[lineId]
		if !got.Equal(want) {
			vErrs.Append(*ft.NewBusinessViolation("parts", ReasonAllocationIncomplete,
				"order line '"+lineId+"' holds "+want.String()+
					" on this bill but the split allocates "+got.String()))
		}
	}
	for lineId := range given {
		if _, present := wanted[lineId]; !present {
			vErrs.Append(*ft.NewBusinessViolation("parts", ReasonAllocationIncomplete,
				"the split allocates order line '"+lineId+
					"', which this bill does not hold"))
		}
	}

	if vErrs.Count() == 0 {
		return nil
	}
	return vErrs
}

// writeSplitParts creates the new bills and apportions the source's allocations across them.
func writeSplitParts(
	ctx corectx.Context,
	source dmodel.DynamicFields,
	existing []dmodel.DynamicFields,
	params SplitBillParams,
	policy SalesPolicy,
) ([]string, decimal.Decimal, error) {
	orgId := stringOf(source, basemodel.FieldOrgId)
	orderId := stringOf(source, models.SalesBillFieldSalesOrderId)
	currency := stringOf(source, models.SalesBillFieldCurrencyCode)

	billIds := make([]string, 0, len(params.Parts))
	for range params.Parts {
		id, err := model.NewId()
		if err != nil {
			return nil, decimal.Zero, err
		}
		billIds = append(billIds, string(*id))
	}

	// Each source allocation is apportioned across the parts in proportion to the quantity each part
	// takes of that line. One allocator call per line, so the residual lands on exactly one part per
	// line rather than being spread and re-rounded.
	perBill := make(map[string]*billTotals, len(billIds))
	for _, billId := range billIds {
		perBill[billId] = &billTotals{}
	}

	for _, allocation := range existing {
		lineId := stringOf(allocation, models.SalesBillLineFieldSalesOrderLineId)

		inputs := make([]AllocationInput, 0, len(params.Parts))
		for index, part := range params.Parts {
			inputs = append(inputs, AllocationInput{
				Key:       billIds[index],
				Reference: part.Allocations[lineId],
				Tiebreak:  int32(index),
			})
		}

		net := AllocateAcrossBills(
			decimalOf(allocation, models.SalesBillLineFieldAllocatedNetAmount),
			inputs, policy.RoundingScale)
		tax := AllocateAcrossBills(
			decimalOf(allocation, models.SalesBillLineFieldAllocatedTaxAmount),
			inputs, policy.RoundingScale)
		total := AllocateAcrossBills(
			decimalOf(allocation, models.SalesBillLineFieldAllocatedTotalAmount),
			inputs, policy.RoundingScale)

		for index, part := range params.Parts {
			billId := billIds[index]
			quantity := part.Allocations[lineId]
			if quantity.IsZero() {
				// A part taking none of this line gets no allocation row for it. A zero row would
				// clutter the bill with lines the customer did not buy.
				continue
			}
			if err := insertBillLine(ctx, billId, lineId, orgId,
				quantity, net[billId], tax[billId], total[billId]); err != nil {
				return nil, decimal.Zero, err
			}
			perBill[billId].add(net[billId], tax[billId], total[billId])
		}
	}

	grandTotal := decimal.Zero
	for index, billId := range billIds {
		totals := perBill[billId]
		if err := insertBill(ctx, billId, orderId, orgId, currency,
			billNumberOf(source, index), totals); err != nil {
			return nil, decimal.Zero, err
		}
		grandTotal = grandTotal.Add(totals.total)
	}
	return billIds, grandTotal, nil
}

// billTotals accumulates one new bill's figures as its allocations are written.
type billTotals struct {
	net   decimal.Decimal
	tax   decimal.Decimal
	total decimal.Decimal
}

func (this *billTotals) add(net, tax, total decimal.Decimal) {
	this.net = this.net.Add(net)
	this.tax = this.tax.Add(tax)
	this.total = this.total.Add(total)
}

// billNumberOf derives a child bill's number from its parent's, rather than allocating a sequential
// one, so BILL-7 splits into BILL-7-1 and BILL-7-2 and a customer holding one can be matched to the
// original without a lookup.
func billNumberOf(source dmodel.DynamicFields, index int) string {
	parent := stringOf(source, models.SalesBillFieldBillNumber)
	return parent + "-" + decimal.NewFromInt(int64(index+1)).String()
}

func insertBill(
	ctx corectx.Context, billId, orderId, orgId, currency, billNumber string, totals *billTotals,
) error {
	engine, err := engineFor(models.SalesBillSchemaName)
	if err != nil {
		return err
	}
	_, err = engine.ResourceRepository().Insert(ctx, dmodel.DynamicFields{
		models.SalesBillFieldId:            billId,
		models.SalesBillFieldBillNumber:    billNumber,
		models.SalesBillFieldSalesOrderId:  orderId,
		models.SalesBillFieldStatus:        string(models.SalesBillStatusOpen),
		models.SalesBillFieldPaymentStatus: string(models.SalesOrderPaymentStatusUnpaid),
		models.SalesBillFieldCurrencyCode:  currency,
		models.SalesBillFieldSubtotal:      totals.net,
		models.SalesBillFieldDiscountTotal: decimal.Zero,
		models.SalesBillFieldTaxTotal:      totals.tax,
		models.SalesBillFieldTotalAmount:   totals.total,
		basemodel.FieldOrgId:               orgId,
	})
	return err
}

func insertBillLine(
	ctx corectx.Context, billId, orderLineId, orgId string,
	quantity, net, tax, total decimal.Decimal,
) error {
	engine, err := engineFor(models.SalesBillLineSchemaName)
	if err != nil {
		return err
	}
	id, err := model.NewId()
	if err != nil {
		return err
	}
	_, err = engine.ResourceRepository().Insert(ctx, dmodel.DynamicFields{
		models.SalesBillLineFieldId:                   string(*id),
		models.SalesBillLineFieldSalesBillId:          billId,
		models.SalesBillLineFieldSalesOrderLineId:     orderLineId,
		models.SalesBillLineFieldQuantity:             quantity,
		models.SalesBillLineFieldAllocatedNetAmount:   net,
		models.SalesBillLineFieldAllocatedTaxAmount:   tax,
		models.SalesBillLineFieldAllocatedTotalAmount: total,
		basemodel.FieldOrgId:                          orgId,
	})
	return err
}

// cancelSupersededBill marks a bill superseded and writes the lineage rows. Never a delete: a
// payment already recorded against the source must still resolve to something, and an auditor
// tracing it needs a row that explains where the value went.
func cancelSupersededBill(
	ctx corectx.Context, source dmodel.DynamicFields, targets []string, relationType string,
) error {
	sourceId := stringOf(source, models.SalesBillFieldId)
	orgId := stringOf(source, basemodel.FieldOrgId)

	billEngine, err := engineFor(models.SalesBillSchemaName)
	if err != nil {
		return err
	}
	if _, err := billEngine.ResourceRepository().Update(ctx, dmodel.DynamicFields{
		models.SalesBillFieldId:          sourceId,
		models.SalesBillFieldStatus:      string(models.SalesBillStatusCancelled),
		models.SalesBillFieldCancelledAt: model.ModelDateTime(time.Now().UTC()),
	}); err != nil {
		return err
	}

	relationEngine, err := engineFor(models.SalesBillRelationSchemaName)
	if err != nil {
		return err
	}
	for _, targetId := range targets {
		id, err := model.NewId()
		if err != nil {
			return err
		}
		if _, err := relationEngine.ResourceRepository().Insert(ctx, dmodel.DynamicFields{
			models.SalesBillRelationFieldId:           string(*id),
			models.SalesBillRelationFieldSourceBillId: sourceId,
			models.SalesBillRelationFieldTargetBillId: targetId,
			models.SalesBillRelationFieldType:         relationType,
			basemodel.FieldOrgId:                      orgId,
		}); err != nil {
			return err
		}
	}
	return nil
}
