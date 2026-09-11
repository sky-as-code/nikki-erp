package services

import (
	"time"

	"github.com/pkg/errors"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	lock "github.com/sky-as-code/nikki-erp/modules/core/infra/distributedlock"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// Creating a return. A return is a new document, never an edit of the order: the order is what the
// customer was charged against and what any fiscal document describes, so editing it would leave
// the invoice and the order disagreeing about what was sold.
//
// This operation moves no goods and no money. Everything irreversible belongs to ProcessReturn, a
// separate action with a separate permission.

const (
	ReasonReturnOrderNotFound = "sales_return.order_not_found"
	ReasonReturnOrderNotSold  = "sales_return.order_not_sold"
	ReasonReturnNoLines       = "sales_return.no_lines"
	ReasonReturnLineNotFound  = "sales_return.line_not_found"
	ReasonReturnLineForeign   = "sales_return.line_belongs_to_another_order"
	ReasonReturnLineDuplicate = "sales_return.line_repeated"
	ReasonReturnQuantityRange = "sales_return.quantity_must_be_positive"

	// ReasonReturnQuantityExceedsReturnable stops the same goods being refunded twice. Named after the
	// business rule rather than the arithmetic, because an operator needs to know what to do.
	ReasonReturnQuantityExceedsReturnable = "sales_return.quantity_exceeds_returnable"

	// ReasonReturnWindowClosed refuses a return the policy no longer allows.
	ReasonReturnWindowClosed = "sales_return.return_window_closed"

	// ReasonReturnComboPartial refuses returning one component of a combo priced as a unit: a single
	// component has no independent price to refund at.
	ReasonReturnComboPartial = "sales_return.combo_must_return_entire"
)

// CreateReturnLine names one order line and how much is coming back. No refund amount: it is
// COMPUTED from what the line historically carried, since accepting one from the caller would be a
// way to take money out of the business by asking for it.
type CreateReturnLine struct {
	SalesOrderLineId string
	Quantity         decimal.Decimal

	// FulfillmentId and FulfillmentItemId link this line to the delivery it answers for, when it
	// came from one. Both empty for an ordinary goods return, where the customer simply brought
	// something back and no machine was involved.
	FulfillmentId     string
	FulfillmentItemId string

	// RequestedQty is how much to refund, which is NOT always the goods quantity: a refund-only line
	// asks for money against goods that never arrived, so it carries a requested amount while the
	// goods quantity stays zero. Zero here means "same as Quantity", which is what an ordinary
	// return means.
	RequestedQty decimal.Decimal
}

type CreateReturnParams struct {
	SalesOrderId         string
	Reason               string
	InventoryDisposition string
	Lines                []CreateReturnLine

	// RefundReason categorises why the money is going back, for the paths that act on it. Empty
	// defaults to customer_requested, which is what a person raising a return means.
	RefundReason models.SalesRefundReason

	// ReturnType decides whether the inventory step runs. Empty defaults to goods_return; a failed
	// dispense passes refund_only, because the goods never reached the customer and Inventory has
	// already dealt with them.
	ReturnType models.SalesReturnType
}

type CreateReturnResult struct {
	SalesReturnId string
	ReturnNumber  string
	Status        string

	// RefundTotal is reported here so the agent raising the return can tell the customer before
	// anything is committed.
	RefundTotal decimal.Decimal

	// A services-only return needs no inventory movement, and an unpaid order owes no refund.
	InventoryReturnStatus  string
	RefundStatus           string
	FiscalAdjustmentStatus string

	Lines []CreateReturnResultLine
}

type CreateReturnResultLine struct {
	SalesOrderLineId        string
	Quantity                decimal.Decimal
	RefundAmount            decimal.Decimal
	RefundTaxAmount         decimal.Decimal
	RequiresInventoryReturn bool

	// Carried through from the request so the write has everything in one place. See
	// CreateReturnLine for what each means.
	FulfillmentId     string
	FulfillmentItemId string
	RequestedQty      decimal.Decimal
}

// CreateReturn takes the ORDER's lock, not the return's: the decision reads every line's returned
// quantity and every existing return against the order, and two concurrent returns of the same
// line would each see the other as unclaimed.
//
// refuseReturn builds a single-violation refusal, the shape every gate in this file returns.
func refuseReturn(field, reason, message string) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(field, reason, message))
	return vErrs
}

func CreateReturn(
	ctx corectx.Context,
	params CreateReturnParams,
	dLock lock.DistributedLock,
	policy SalesPolicy,
) (*CreateReturnResult, *ft.ClientErrors, error) {
	if dLock == nil {
		// Refusing to run unguarded: two concurrent returns of the same line would each see the
		// other's quantity as unclaimed and refund the customer twice.
		return nil, nil, errors.New(
			"the distributed lock is not available; a return cannot be created unguarded")
	}

	order, err := loadRecord(ctx, models.SalesOrderSchemaName, models.SalesOrderFieldId, params.SalesOrderId)
	if err != nil {
		return nil, nil, err
	}
	if order == nil {
		return nil, refuseReturn(models.SalesReturnFieldSalesOrderId, ReasonReturnOrderNotFound,
			"no such order"), nil
	}

	key := confirmLockKeyOf(params.SalesOrderId)
	acquired, err := dLock.AcquireWithRetry(ctx, key, confirmLockTtl, confirmLockRetryCount, confirmLockRetryDelay)
	if err != nil {
		return nil, nil, err
	}
	if !acquired {
		return nil, refuseReturn(models.SalesReturnFieldSalesOrderId, ReasonLockUnavailable,
			"this order is being changed by another request; try again"), nil
	}
	defer func() { _ = dLock.Release(ctx, key) }()

	// Re-read under the lock: the first read described a world that may have moved.
	order, err = loadRecord(ctx, models.SalesOrderSchemaName, models.SalesOrderFieldId, params.SalesOrderId)
	if err != nil {
		return nil, nil, err
	}
	if order == nil {
		return nil, refuseReturn(models.SalesReturnFieldSalesOrderId, ReasonReturnOrderNotFound,
			"no such order"), nil
	}

	return createReturnUnderLock(ctx, params, order, policy)
}

func createReturnUnderLock(
	ctx corectx.Context,
	params CreateReturnParams,
	order dmodel.DynamicFields,
	policy SalesPolicy,
) (*CreateReturnResult, *ft.ClientErrors, error) {
	if vErrs := assertOrderReturnable(order); vErrs != nil {
		return nil, vErrs, nil
	}
	if len(params.Lines) == 0 {
		return nil, refuseReturn("lines", ReasonReturnNoLines,
			"a return must name at least one line"), nil
	}

	priced, vErrs, err := priceReturnLines(ctx, params, order, policy)
	if err != nil {
		return nil, nil, err
	}
	if vErrs != nil {
		return nil, vErrs, nil
	}

	refundTotal := decimal.Zero
	for _, line := range priced {
		refundTotal = refundTotal.Add(line.RefundAmount)
	}

	returnable := make([]ReturnableLine, 0, len(priced))
	for _, line := range priced {
		returnable = append(returnable, ReturnableLine{RequiresFulfillment: line.RequiresInventoryReturn})
	}

	result := &CreateReturnResult{
		Status:      string(models.SalesReturnStatusDraft),
		RefundTotal: refundTotal,
		Lines:       priced,

		InventoryReturnStatus: DeriveInventoryStepStatus(returnable),
		RefundStatus:          DeriveRefundStepStatus(refundTotal),

		// The fiscal step resolves when the return is processed: whether an adjustment is owed
		// depends on whether an invoice was ever issued.
		FiscalAdjustmentStatus: string(models.SalesReturnStepPending),
	}

	if err := writeReturn(ctx, params, order, result); err != nil {
		return nil, nil, err
	}
	return result, nil, nil
}

// assertOrderReturnable refuses an order that never became a sale: neither a draft nor a cancelled
// one has goods with the customer or money in the business.
func assertOrderReturnable(order dmodel.DynamicFields) *ft.ClientErrors {
	status := stringOf(order, models.SalesOrderFieldStatus)
	switch status {
	case string(models.SalesOrderStatusConfirmed),
		string(models.SalesOrderStatusProcessing),
		string(models.SalesOrderStatusCompleted):
		return nil
	}
	return refuseReturn(models.SalesReturnFieldSalesOrderId, ReasonReturnOrderNotSold,
		"only a confirmed, processing or completed order can be returned; this one is "+status)
}

// priceReturnLines: the quantity must be within what is still returnable, and the refund comes
// from the amounts the line HISTORICALLY carried, so a line discounted at the time comes back at
// its discounted share.
func priceReturnLines(
	ctx corectx.Context,
	params CreateReturnParams,
	order dmodel.DynamicFields,
	policy SalesPolicy,
) ([]CreateReturnResultLine, *ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()
	priced := make([]CreateReturnResultLine, 0, len(params.Lines))
	seen := make(map[string]bool, len(params.Lines))
	refundOnly := returnTypeOf(params) == models.SalesReturnTypeRefundOnly

	for _, requested := range params.Lines {
		if seen[requested.SalesOrderLineId] {
			vErrs.Append(*ft.NewBusinessViolation(models.SalesReturnLineFieldSalesOrderLineId,
				ReasonReturnLineDuplicate,
				"line "+requested.SalesOrderLineId+" is named twice; combine the quantities"))
			continue
		}
		seen[requested.SalesOrderLineId] = true

		// A refund-only line asks for money against goods that never arrived, so its GOODS quantity
		// is legitimately zero; what must be positive there is the requested refund quantity.
		claimed := requested.Quantity
		if refundOnly && claimed.IsZero() {
			claimed = requested.RequestedQty
		}
		if !claimed.IsPositive() {
			vErrs.Append(*ft.NewBusinessViolation(models.SalesReturnLineFieldQuantity,
				ReasonReturnQuantityRange,
				"the quantity to return must be greater than zero"))
			continue
		}

		line, err := loadRecord(ctx, models.SalesOrderLineSchemaName,
			models.SalesOrderLineFieldId, requested.SalesOrderLineId)
		if err != nil {
			return nil, nil, err
		}
		if line == nil {
			vErrs.Append(*ft.NewBusinessViolation(models.SalesReturnLineFieldSalesOrderLineId,
				ReasonReturnLineNotFound, "no such order line: "+requested.SalesOrderLineId))
			continue
		}
		if stringOf(line, models.SalesOrderLineFieldSalesOrderId) != stringOf(order, models.SalesOrderFieldId) {
			vErrs.Append(*ft.NewBusinessViolation(models.SalesReturnLineFieldSalesOrderLineId,
				ReasonReturnLineForeign,
				"line "+requested.SalesOrderLineId+" belongs to a different order"))
			continue
		}

		previouslyReturned, err := returnedQuantityOf(ctx, requested.SalesOrderLineId, line)
		if err != nil {
			return nil, nil, err
		}

		basis := ReturnableLine{
			Ordered:             decimalOf(line, models.SalesOrderLineFieldOrderedQuantity),
			Fulfilled:           decimalOf(line, models.SalesOrderLineFieldFulfilledQuantity),
			PreviouslyReturned:  previouslyReturned,
			RequiresFulfillment: boolOrTrue(line, models.SalesOrderLineFieldRequiresFulfillment),
		}
		allowed := ReturnableQuantity(basis)

		if requested.Quantity.GreaterThan(allowed) {
			vErrs.Append(*ft.NewBusinessViolation(models.SalesReturnLineFieldQuantity,
				ReasonReturnQuantityExceedsReturnable,
				"line "+requested.SalesOrderLineId+" has "+allowed.String()+
					" returnable, and "+requested.Quantity.String()+" was asked for"))
			continue
		}

		refund, tax := refundAmountFor(line, requested.Quantity, policy.RoundingScale)
		// A refund-only line never asks Inventory for anything, whatever the product's basis says:
		// the goods never reached the customer, so there is nothing for them to bring back.
		requiresInventoryReturn := RequiresInventoryReturn(basis) && !refundOnly

		requestedQty := requested.RequestedQty
		if requestedQty.IsZero() {
			requestedQty = requested.Quantity
		}

		priced = append(priced, CreateReturnResultLine{
			SalesOrderLineId:        requested.SalesOrderLineId,
			Quantity:                requested.Quantity,
			RefundAmount:            refund,
			RefundTaxAmount:         tax,
			RequiresInventoryReturn: requiresInventoryReturn,
			FulfillmentId:           requested.FulfillmentId,
			FulfillmentItemId:       requested.FulfillmentItemId,
			RequestedQty:            requestedQty,
		})
	}

	if vErrs.Count() > 0 {
		return nil, vErrs, nil
	}
	return priced, nil, nil
}

// refundAmountFor is historical, never current: the line's own final and tax amounts are what the
// customer was actually charged, including any discount or voucher share. A 60,000 line carrying
// 12,000 of a 20,000 voucher refunds 48,000. Pro-rated by quantity for a partial return.
func refundAmountFor(
	line dmodel.DynamicFields, quantity decimal.Decimal, scale int32,
) (decimal.Decimal, decimal.Decimal) {
	ordered := decimalOf(line, models.SalesOrderLineFieldOrderedQuantity)
	if !ordered.IsPositive() {
		return decimal.Zero, decimal.Zero
	}

	share := quantity.Div(ordered)
	final := decimalOf(line, models.SalesOrderLineFieldFinalAmount)
	tax := decimalOf(line, models.SalesOrderLineFieldTaxAmount)

	return final.Mul(share).Round(scale), tax.Mul(share).Round(scale)
}

// returnedQuantityOf counts return lines of ACCEPTED returns, not the order line's own
// returned_quantity column, which only moves once goods physically arrive: counting only completed
// movements would let two returns each claim the same quantity while the first is in flight.
// Cancelled returns release their claim, which is what makes cancelling useful.
func returnedQuantityOf(
	ctx corectx.Context, orderLineId string, line dmodel.DynamicFields,
) (decimal.Decimal, error) {
	claimed, err := searchBy(ctx, models.SalesReturnLineSchemaName,
		models.SalesReturnLineFieldSalesOrderLineId, orderLineId)
	if err != nil {
		return decimal.Zero, err
	}

	total := decimal.Zero
	for _, returnLine := range claimed {
		returnId := stringOf(returnLine, models.SalesReturnLineFieldSalesReturnId)
		parent, err := loadRecord(ctx, models.SalesReturnSchemaName, models.SalesReturnFieldId, returnId)
		if err != nil {
			return decimal.Zero, err
		}
		if parent == nil {
			continue
		}
		if stringOf(parent, models.SalesReturnFieldStatus) == string(models.SalesReturnStatusCancelled) {
			continue
		}
		total = total.Add(decimalOf(returnLine, models.SalesReturnLineFieldQuantity))
	}

	// The order line's own column is the floor, not the answer: it records goods that have physically
	// come back, which a return in flight has not. The larger of the two means neither a stale column
	// nor a missing return line can understate what is claimed.
	if settled := decimalOf(line, models.SalesOrderLineFieldReturnedQuantity); settled.GreaterThan(total) {
		return settled, nil
	}
	return total, nil
}

// writeReturn uses one transaction: a return with no lines would look like a recorded customer
// request while claiming none of the quantity it should reserve.
func writeReturn(
	ctx corectx.Context,
	params CreateReturnParams,
	order dmodel.DynamicFields,
	result *CreateReturnResult,
) error {
	id, err := model.NewId()
	if err != nil {
		return err
	}
	returnId := string(*id)
	orgId := stringOf(order, basemodel.FieldOrgId)

	number, err := model.NewId()
	if err != nil {
		return err
	}
	returnNumber := "RET-" + string(*number)

	err = withTransaction(ctx, models.SalesReturnSchemaName, func(tranxCtx corectx.Context) error {
		engineRepo, err := repoFor(models.SalesReturnSchemaName)
		if err != nil {
			return err
		}
		if _, err := engineRepo.Insert(tranxCtx, dmodel.DynamicFields{
			models.SalesReturnFieldId:                     returnId,
			models.SalesReturnFieldOrgId:                  orgId,
			models.SalesReturnFieldReturnNumber:           returnNumber,
			models.SalesReturnFieldSalesOrderId:           stringOf(order, models.SalesOrderFieldId),
			models.SalesReturnFieldStatus:                 result.Status,
			models.SalesReturnFieldInventoryReturnStatus:  result.InventoryReturnStatus,
			models.SalesReturnFieldRefundStatus:           result.RefundStatus,
			models.SalesReturnFieldFiscalAdjustmentStatus: result.FiscalAdjustmentStatus,
			models.SalesReturnFieldReason:                 params.Reason,
			models.SalesReturnFieldRefundReason:           string(refundReasonOf(params)),
			models.SalesReturnFieldReturnType:             string(returnTypeOf(params)),
			models.SalesReturnFieldInventoryDisposition:   params.InventoryDisposition,
			models.SalesReturnFieldRefundTotal:            result.RefundTotal,
			models.SalesReturnFieldRequestedAt:            model.ModelDateTime(time.Now().UTC()),
		}); err != nil {
			return err
		}

		lineRepo, err := repoFor(models.SalesReturnLineSchemaName)
		if err != nil {
			return err
		}
		for _, line := range result.Lines {
			lineId, err := model.NewId()
			if err != nil {
				return err
			}
			if _, err := lineRepo.Insert(tranxCtx, dmodel.DynamicFields{
				models.SalesReturnLineFieldId:                      string(*lineId),
				models.SalesReturnLineFieldOrgId:                   orgId,
				models.SalesReturnLineFieldSalesReturnId:           returnId,
				models.SalesReturnLineFieldSalesOrderLineId:        line.SalesOrderLineId,
				models.SalesReturnLineFieldQuantity:                line.Quantity,
				models.SalesReturnLineFieldRequestedQty:            line.RequestedQty,
				models.SalesReturnLineFieldRefundedQty:             decimal.Zero,
				models.SalesReturnLineFieldRefundAmount:            line.RefundAmount,
				models.SalesReturnLineFieldRefundTaxAmount:         line.RefundTaxAmount,
				models.SalesReturnLineFieldRequiresInventoryReturn: line.RequiresInventoryReturn,
				models.SalesReturnLineFieldFulfillmentId:           nullableId(line.FulfillmentId),
				models.SalesReturnLineFieldFulfillmentItemId:       nullableId(line.FulfillmentItemId),
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	result.SalesReturnId = returnId
	result.ReturnNumber = returnNumber
	return nil
}

// refundReasonOf defaults to a customer having asked, which is what a person raising a return means.
// The automatic path states fulfillment_failure explicitly, because it is the only kind of refund
// that may exist with nobody having asked for one.
func refundReasonOf(params CreateReturnParams) models.SalesRefundReason {
	if params.RefundReason == "" {
		return models.SalesRefundReasonCustomerRequested
	}
	return params.RefundReason
}

// returnTypeOf defaults to goods coming back, so an existing caller that knows nothing about
// refund-only returns keeps the behaviour it had.
func returnTypeOf(params CreateReturnParams) models.SalesReturnType {
	if params.ReturnType == "" {
		return models.SalesReturnTypeGoodsReturn
	}
	return params.ReturnType
}

// nullableId writes a real NULL rather than an empty string for an absent optional id, so a query
// filtering on "has a fulfillment" is not fooled by a blank that is neither set nor null.
func nullableId(value string) any {
	if value == "" {
		return nil
	}
	return value
}
