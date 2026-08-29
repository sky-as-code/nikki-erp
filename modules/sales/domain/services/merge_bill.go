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

// Merging bills. Several settlement units become one - the inverse of a split, with the same two
// invariants: the total does not change, and nothing is deleted.
//
// A merge whose conditions fail is REJECTED outright, never partially applied. Every gate runs
// BEFORE anything is written and the write happens in one transaction, because a merge that got
// halfway would leave some bills cancelled and their value on no live bill at all.
//
// The gates: currency must match, because Sales has no FX; bills must belong to one ORDER, because
// the allocation invariant is stated per order; and all must be open.

type MergeBillParams struct {
	SourceBillIds []string
}

type MergeBillResult struct {
	MergedBillId  string
	SourceBillIds []string
	TotalBefore   decimal.Decimal
	TotalAfter    decimal.Decimal
}

const (
	ReasonMergeNeedsTwoBills = "sales_bill.merge_needs_two_bills"
	ReasonDifferentOrders    = "sales_bill.different_orders"
	ReasonMergeTotalChanged  = "sales_bill.merge_total_changed"
)

func MergeBills(
	ctx corectx.Context, params MergeBillParams, dLock lock.DistributedLock,
) (*MergeBillResult, *ft.ClientErrors, error) {
	if dLock == nil {
		return nil, nil, errors.New(
			"the distributed lock is not available; bills cannot be merged without it")
	}
	if len(params.SourceBillIds) < 2 {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("source_bill_ids", ReasonMergeNeedsTwoBills,
			"a merge needs at least two bills"))
		return nil, vErrs, nil
	}

	// The order is resolved from the first bill purely to name the lock. Whether every bill really
	// belongs to it is checked under the lock, where the answer cannot change underneath us.
	first, err := loadRecord(ctx,
		models.SalesBillSchemaName, models.SalesBillFieldId, params.SourceBillIds[0])
	if err != nil {
		return nil, nil, err
	}
	if first == nil {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("source_bill_ids", ReasonBillNotFound,
			"no bill exists with id '"+params.SourceBillIds[0]+"'"))
		return nil, vErrs, nil
	}

	orderId := stringOf(first, models.SalesBillFieldSalesOrderId)
	key := confirmLockKeyOf(orderId)
	acquired, err := dLock.AcquireWithRetry(
		ctx, key, confirmLockTtl, confirmLockRetryCount, confirmLockRetryDelay)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "acquiring the lock of order '%s'", orderId)
	}
	if !acquired {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("source_bill_ids", ReasonLockUnavailable,
			"this order is being changed by another request; try again"))
		return nil, vErrs, nil
	}
	defer func() { _ = dLock.Release(ctx, key) }()

	return mergeUnderLock(ctx, params, orderId)
}

func mergeUnderLock(
	ctx corectx.Context, params MergeBillParams, orderId string,
) (*MergeBillResult, *ft.ClientErrors, error) {
	sources, vErrs, err := loadMergeableBills(ctx, params.SourceBillIds, orderId)
	if err != nil || vErrs != nil {
		return nil, vErrs, err
	}

	totalBefore := decimal.Zero
	for _, source := range sources {
		totalBefore = totalBefore.Add(decimalOf(source, models.SalesBillFieldTotalAmount))
	}

	mergedId, err := model.NewId()
	if err != nil {
		return nil, nil, err
	}
	targetId := string(*mergedId)

	var totalAfter decimal.Decimal
	err = withTransaction(ctx, models.SalesBillSchemaName, func(tranxCtx corectx.Context) error {
		totalAfter, err = writeMergedBill(tranxCtx, targetId, orderId, sources)
		if err != nil {
			return err
		}

		// Each source is cancelled and points at the survivor: one relation row per source, all with
		// the same target, which distinguishes a merge from a split in the lineage.
		for _, source := range sources {
			if err := cancelSupersededBill(tranxCtx, source, []string{targetId},
				string(models.SalesBillRelationMergedInto)); err != nil {
				return err
			}
		}

		vErrs, err := AssertOrderAllocationBalances(tranxCtx, orderId)
		if err != nil {
			return err
		}
		if vErrs != nil {
			return errors.New("the merge left the order's bills unbalanced: " +
				vErrs.ToError().Error())
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	if !totalAfter.Equal(totalBefore) {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("source_bill_ids", ReasonMergeTotalChanged,
			"the merge produced "+totalAfter.String()+" from bills totalling "+
				totalBefore.String()))
		return nil, vErrs, nil
	}

	return &MergeBillResult{
		MergedBillId:  targetId,
		SourceBillIds: params.SourceBillIds,
		TotalBefore:   totalBefore,
		TotalAfter:    totalAfter,
	}, nil, nil
}

// loadMergeableBills runs every gate before anything is written and collects EVERY failure rather
// than stopping at the first, because the merge is rejected as a whole.
func loadMergeableBills(
	ctx corectx.Context, billIds []string, orderId string,
) ([]dmodel.DynamicFields, *ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()
	sources := make([]dmodel.DynamicFields, 0, len(billIds))

	currency := ""
	seen := make(map[string]bool, len(billIds))

	for _, billId := range billIds {
		if seen[billId] {
			// The same bill twice would be cancelled twice and counted twice in the total.
			vErrs.Append(*ft.NewBusinessViolation("source_bill_ids", ReasonMergeNeedsTwoBills,
				"bill '"+billId+"' is listed more than once"))
			continue
		}
		seen[billId] = true

		record, err := loadRecord(ctx, models.SalesBillSchemaName, models.SalesBillFieldId, billId)
		if err != nil {
			return nil, nil, err
		}
		if record == nil {
			vErrs.Append(*ft.NewBusinessViolation("source_bill_ids", ReasonBillNotFound,
				"no bill exists with id '"+billId+"'"))
			continue
		}

		if !models.NewSalesBillFrom(record).IsOpen() {
			vErrs.Append(*ft.NewBusinessViolation("source_bill_ids", ReasonBillNotOpen,
				"bill '"+billId+"' is '"+stringOf(record, models.SalesBillFieldStatus)+
					"' and cannot be merged"))
			continue
		}

		if stringOf(record, models.SalesBillFieldSalesOrderId) != orderId {
			// The allocation invariant is per order, so a cross-order merge would break it on both.
			vErrs.Append(*ft.NewBusinessViolation("source_bill_ids", ReasonDifferentOrders,
				"bill '"+billId+"' belongs to a different sales order"))
			continue
		}

		billCurrency := stringOf(record, models.SalesBillFieldCurrencyCode)
		if currency == "" {
			currency = billCurrency
		} else if billCurrency != currency {
			// Refused rather than converted: Sales has no FX, and inventing a rate would be a pricing
			// decision nobody authorised.
			vErrs.Append(*ft.NewBusinessViolation("source_bill_ids", ReasonCurrencyMismatch,
				"bill '"+billId+"' is in "+billCurrency+" but the others are in "+currency))
			continue
		}

		sources = append(sources, record)
	}

	if vErrs.Count() > 0 {
		return nil, vErrs, nil
	}
	return sources, nil, nil
}

// writeMergedBill combines allocations of the same order line from different bills into one row,
// because the unique on (sales_bill_id, sales_order_line_id) permits only one per line per bill.
func writeMergedBill(
	ctx corectx.Context, targetId, orderId string, sources []dmodel.DynamicFields,
) (decimal.Decimal, error) {
	orgId := stringOf(sources[0], basemodel.FieldOrgId)
	currency := stringOf(sources[0], models.SalesBillFieldCurrencyCode)

	type combined struct {
		quantity decimal.Decimal
		net      decimal.Decimal
		tax      decimal.Decimal
		total    decimal.Decimal
	}
	byLine := map[string]*combined{}
	order := make([]string, 0)

	for _, source := range sources {
		allocations, err := searchBy(ctx, models.SalesBillLineSchemaName,
			models.SalesBillLineFieldSalesBillId, stringOf(source, models.SalesBillFieldId))
		if err != nil {
			return decimal.Zero, err
		}
		for _, allocation := range allocations {
			lineId := stringOf(allocation, models.SalesBillLineFieldSalesOrderLineId)
			if _, present := byLine[lineId]; !present {
				byLine[lineId] = &combined{}
				order = append(order, lineId)
			}
			entry := byLine[lineId]
			entry.quantity = entry.quantity.Add(
				decimalOf(allocation, models.SalesBillLineFieldQuantity))
			entry.net = entry.net.Add(
				decimalOf(allocation, models.SalesBillLineFieldAllocatedNetAmount))
			entry.tax = entry.tax.Add(
				decimalOf(allocation, models.SalesBillLineFieldAllocatedTaxAmount))
			entry.total = entry.total.Add(
				decimalOf(allocation, models.SalesBillLineFieldAllocatedTotalAmount))
		}
	}

	totals := &billTotals{}
	for _, lineId := range order {
		entry := byLine[lineId]
		if err := insertBillLine(ctx, targetId, lineId, orgId,
			entry.quantity, entry.net, entry.tax, entry.total); err != nil {
			return decimal.Zero, err
		}
		totals.add(entry.net, entry.tax, entry.total)
	}

	// Summed, never re-allocated. A merge combines amounts that were already apportioned exactly, so
	// re-running the allocator would re-round figures that already balance and could lose a dong.
	if err := insertBill(ctx, targetId, orderId, orgId, currency,
		mergedBillNumberOf(sources), totals); err != nil {
		return decimal.Zero, err
	}
	return totals.total, nil
}

// mergedBillNumberOf names the survivor after the first source it consumed, so a customer holding
// one of the old bills can be matched to the new one without a lookup.
func mergedBillNumberOf(sources []dmodel.DynamicFields) string {
	return stringOf(sources[0], models.SalesBillFieldBillNumber) + "-M" +
		decimal.NewFromInt(time.Now().UTC().Unix()).String()
}
