package services

import (
	"fmt"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// Writing bills and their allocations.
//
// One writer for every path that creates a bill - the initial bill raised at confirmation, and the
// bills a split or a merge produces. They differ in how the allocations are computed, never in what
// a bill row looks like, so the shape lives here and each caller brings its own figures.
//
// None of these manage a transaction. Every caller already runs inside one, because a bill without
// its lines is not a half-written bill, it is a bill that claims to settle nothing.

// billTotals accumulates one bill's figures as its allocations are written.
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

func insertBill(
	ctx corectx.Context, billId, orderId, orgId, currency, billNumber string, totals *billTotals,
) error {
	engineRepo, err := repoFor(models.SalesBillSchemaName)
	if err != nil {
		return err
	}
	result, err := engineRepo.Insert(ctx, dmodel.DynamicFields{
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
		// Direct repository inserts bypass the resource service's model defaults.
		basemodel.FieldIsArchived: false,
	})
	return billInsertError(result, err)
}

func insertBillLine(
	ctx corectx.Context, billId, orderLineId, orgId string,
	quantity, net, tax, total decimal.Decimal,
) error {
	engineRepo, err := repoFor(models.SalesBillLineSchemaName)
	if err != nil {
		return err
	}
	id, err := model.NewId()
	if err != nil {
		return err
	}
	result, err := engineRepo.Insert(ctx, dmodel.DynamicFields{
		models.SalesBillLineFieldId:                   string(*id),
		models.SalesBillLineFieldSalesBillId:          billId,
		models.SalesBillLineFieldSalesOrderLineId:     orderLineId,
		models.SalesBillLineFieldQuantity:             quantity,
		models.SalesBillLineFieldAllocatedNetAmount:   net,
		models.SalesBillLineFieldAllocatedTaxAmount:   tax,
		models.SalesBillLineFieldAllocatedTotalAmount: total,
		basemodel.FieldOrgId:                          orgId,
	})
	return billInsertError(result, err)
}

// Repository constraint failures can arrive as ClientErrors with a nil Go error.
// Return an error so the enclosing transaction stops and rolls back.
func billInsertError(result *dyn.OpResult[int], err error) error {
	if err != nil {
		return err
	}
	if result == nil {
		return fmt.Errorf("bill insert returned no result")
	}
	if err := result.ClientErrors.ToError(); err != nil {
		return err
	}
	if !result.HasData || result.Data != 1 {
		return fmt.Errorf("bill insert did not create one row")
	}
	return nil
}

type billAllocation struct {
	lineId                    string
	quantity, net, tax, total decimal.Decimal
}

func insertBillAllocations(ctx corectx.Context, billId, orgId string, allocations []billAllocation) error {
	for _, entry := range allocations {
		if err := insertBillLine(ctx, billId, entry.lineId, orgId,
			entry.quantity, entry.net, entry.tax, entry.total); err != nil {
			return err
		}
	}
	return nil
}
