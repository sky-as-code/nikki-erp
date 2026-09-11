package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/distributedlock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
	itInvoicing "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external/invoicing"
	itReturns "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/returns"
)

// The return's authorized surface.
//
// Three permissions, deliberately distinct: recording a request is not the power to refund, and
// processing moves money out of the business, so it carries its own code rather than riding on
// update.

type SalesReturnApplicationServiceImpl struct {
	composable.CrudApplicationService

	orderLock               distributedlock.DistributedLock
	effectiveSettings       itExt.EffectiveSettingsExtService
	fulfillmentReservations itExt.FulfillmentReservationExtService
	paymentOrders           itExt.PaymentOrderExtService
	orderFulfillment        itExt.FulfillmentExtService

	// Optional: no e-invoicing adapter ships in every deployment, and a return still processes
	// without one -- the fiscal document is simply not reversed.
	invoicingProvider itInvoicing.InvoicingExtService
}

func NewSalesReturnApplicationService(
	base composable.CrudApplicationService,
	dLock distributedlock.DistributedLock,
	settings itExt.EffectiveSettingsExtService,
	reservations itExt.FulfillmentReservationExtService,
	orders itExt.PaymentOrderExtService,
	fulfillment itExt.FulfillmentExtService,
	invoicing itInvoicing.InvoicingExtService,
) itReturns.SalesReturnApplicationService {
	return &SalesReturnApplicationServiceImpl{
		CrudApplicationService:  base,
		orderLock:               dLock,
		effectiveSettings:       settings,
		fulfillmentReservations: reservations,
		paymentOrders:           orders,
		orderFulfillment:        fulfillment,
		invoicingProvider:       invoicing,
	}
}

// CreateReturn is collection-level: the return does not exist yet, so there is no record to place
// in an org.
func (this *SalesReturnApplicationServiceImpl) CreateReturn(
	ctx corectx.Context, cmd itReturns.ReturnActionCommand,
) (*dyn.OpResult[any], error) {
	if _, cErrs := this.AssertAction(ctx, PermissionCreateReturn, cmd); cErrs != nil {
		return anyFailure(cErrs, nil)
	}
	return this.runCreateReturn(ctx, cmd)
}

func (this *SalesReturnApplicationServiceImpl) Process(
	ctx corectx.Context, cmd itReturns.ReturnActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionProcessReturn, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runProcessReturn(ctx, cmd)
}

func (this *SalesReturnApplicationServiceImpl) Cancel(
	ctx corectx.Context, cmd itReturns.ReturnActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionCancelReturn, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runCancelReturn(ctx, cmd)
}

func NewSalesReturnLineApplicationService(base composable.CrudApplicationService) itReturns.SalesReturnLineApplicationService {
	return &SalesReturnLineApplicationServiceImpl{CrudApplicationService: base}
}

type SalesReturnLineApplicationServiceImpl struct {
	composable.CrudApplicationService
}

func NewSalesRefundPaymentApplicationService(base composable.CrudApplicationService) itReturns.SalesRefundPaymentApplicationService {
	return &SalesRefundPaymentApplicationServiceImpl{CrudApplicationService: base}
}

type SalesRefundPaymentApplicationServiceImpl struct {
	composable.CrudApplicationService
}

// The moved action bodies follow.
//
// Each success path sets HasData. The engine actions did not, and both REST layers answer 400
// when !HasData (composable/rest_custom.go:49, and engine/restapi.go:225 before it), so
// create_return, process and cancel answered 400 even when they had succeeded and written the
// return. Corrected here on the API owner's decision: the endpoints now answer 200 with the
// result the body already carried.

func (this *SalesReturnApplicationServiceImpl) runCreateReturn(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	policy := services.ResolveSalesPolicy(ctx, this.effectiveSettings)
	result, vErrs, err := services.CreateReturn(ctx, services.CreateReturnParams{
		SalesOrderId:         readStringParam(params, "sales_order_id"),
		Reason:               readStringParam(params, "reason"),
		InventoryDisposition: readStringParam(params, "inventory_disposition"),
		Lines:                readReturnLines(params),
	}, this.orderLock, policy)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
	}
	return &dyn.OpResult[any]{Data: result, HasData: true}, nil
}

func (this *SalesReturnApplicationServiceImpl) runProcessReturn(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	result, vErrs, err := services.ProcessReturn(ctx, services.ProcessReturnParams{
		SalesReturnId: readStringParam(params, paramRecordId),
	}, this.orderLock, this.orderFulfillment, this.invoicingProvider, this.paymentOrders)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
	}
	return &dyn.OpResult[any]{Data: result, HasData: true}, nil
}

func (this *SalesReturnApplicationServiceImpl) runCancelReturn(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	result, vErrs, err := services.CancelReturn(ctx, services.CancelReturnParams{
		SalesReturnId: readStringParam(params, paramRecordId),
		Reason:        readStringParam(params, "reason"),
	}, this.orderLock)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
	}
	return &dyn.OpResult[any]{Data: result, HasData: true}, nil
}

// readReturnLines reads only an order line and a quantity. The refund amount is computed from what
// the line historically carried, never accepted from the caller.
func readReturnLines(params map[string]any) []services.CreateReturnLine {
	raw, ok := params["lines"].([]any)
	if !ok {
		return nil
	}

	lines := make([]services.CreateReturnLine, 0, len(raw))
	for _, item := range raw {
		fields, ok := item.(map[string]any)
		if !ok {
			continue
		}
		lines = append(lines, services.CreateReturnLine{
			SalesOrderLineId: readStringParam(fields, "sales_order_line_id"),
			Quantity:         readDecimalParam(fields, "quantity"),
		})
	}
	return lines
}
