package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	lock "github.com/sky-as-code/nikki-erp/modules/core/infra/distributedlock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"

	c "github.com/sky-as-code/nikki-erp/modules/sales/constants"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
	it "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/order"
)

// The in-process selling port, for a module inside this binary that sells on Sales' behalf.
//
// Authorization happens here, exactly as it does for a REST caller: a kiosk holds its own service
// principal and its own entitlements, and being internal earns it nothing. That is the whole reason
// this lives in app/ rather than being a thin wrapper somebody could call from a domain service.
//
// Everything below delegates to the same domain operations the engine actions use. A second code
// path for kiosks would eventually price, reserve or refund differently from every other channel,
// which is precisely the divergence this port exists to prevent.

type SalesOrderExtServiceImpl struct {
	dLock lock.DistributedLock

	// settings resolves the org's sales policy, which decides expiry windows, operation types and
	// the rest of what pricing and fulfillment read.
	settings itExt.EffectiveSettingsExtService

	tax      itExt.TaxCalculationExtService
	products itExt.ProductVariantExtService
	basis    itExt.ProductPricingBasisExtService

	fulfillment  itExt.FulfillmentExtService
	reservations itExt.FulfillmentReservationExtService
}

func NewSalesOrderExtService(
	dLock lock.DistributedLock,
	settings itExt.EffectiveSettingsExtService,
	tax itExt.TaxCalculationExtService,
	products itExt.ProductVariantExtService,
	basis itExt.ProductPricingBasisExtService,
	fulfillment itExt.FulfillmentExtService,
	reservations itExt.FulfillmentReservationExtService,
) it.SalesOrderExtService {
	return &SalesOrderExtServiceImpl{
		dLock:        dLock,
		settings:     settings,
		tax:          tax,
		products:     products,
		basis:        basis,
		fulfillment:  fulfillment,
		reservations: reservations,
	}
}

func (this *SalesOrderExtServiceImpl) CreateOrder(
	ctx corectx.Context, command it.CreateSalesOrderCommand,
) (*it.CreateSalesOrderResult, error) {
	if cErrs := assertPermission(
		ctx, composable.PermissionCreate, c.SalesOrderResource, c.ResourceScopeOrg); cErrs != nil {
		return &it.CreateSalesOrderResult{ClientErrors: *cErrs}, nil
	}

	lines := make([]services.CreateOrderLine, 0, len(command.Lines))
	for _, line := range command.Lines {
		lines = append(lines, services.CreateOrderLine{
			ProductVariantId: line.ProductVariantId,
			UomId:            line.UomId,
			Quantity:         line.Quantity,
			EstimatedPrice:   line.EstimatedPrice,
		})
	}

	policy := services.ResolveSalesPolicy(ctx, this.settings)
	created, vErrs, err := services.CreateOrder(ctx, services.CreateOrderParams{
		SalesPointId:      command.SalesPointId,
		CurrencyCode:      command.CurrencyCode,
		IdempotencyKey:    command.IdempotencyKey,
		ExternalReference: command.ExternalReference,
		Lines:             lines,

		EstimatedTotalPrice: command.EstimatedTotalPrice,

		// Only the target is passed through. The METHOD is never accepted from a caller: it is
		// resolved from the point's and channel's defaults, so a seller cannot pick the policy that
		// decides its own refund rules.
		Fulfillment: services.CreateOrderFulfillment{
			TargetOutletId: command.TargetOutletId,
		},
	}, this.tax, this.products, this.basis, policy)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &it.CreateSalesOrderResult{ClientErrors: *vErrs}, nil
	}

	return &it.CreateSalesOrderResult{
		HasData: true,
		Data: it.SalesOrderData{
			SalesOrderId:   created.SalesOrderId,
			OrderNumber:    created.OrderNumber,
			SalesChannelId: created.SalesChannelId,
			AlreadyExisted: created.AlreadyExisted,
		},
	}, nil
}

func (this *SalesOrderExtServiceImpl) ConfirmOrder(
	ctx corectx.Context, command it.SalesOrderCommand,
) (*it.ConfirmSalesOrderResult, error) {
	if cErrs := assertPermission(
		ctx, composable.PermissionUpdate, c.SalesOrderResource, c.ResourceScopeOrg); cErrs != nil {
		return &it.ConfirmSalesOrderResult{ClientErrors: *cErrs}, nil
	}

	policy := services.ResolveSalesPolicy(ctx, this.settings)
	confirmed, vErrs, err := services.ConfirmOrder(ctx, command.SalesOrderId,
		this.dLock, this.tax, this.fulfillment, this.basis, policy,
		salesFulfillmentMethodService(), this.reservations)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &it.ConfirmSalesOrderResult{ClientErrors: *vErrs}, nil
	}

	data := it.ConfirmedOrderData{
		SalesOrderId:     confirmed.SalesOrderId,
		Status:           confirmed.Status,
		InitialBillId:    confirmed.InitialBillId,
		AlreadyConfirmed: confirmed.AlreadyConfirmed,
		Pending:          confirmed.Pending,
	}
	if confirmed.KioskFulfillment != nil {
		data.FulfillmentId = confirmed.KioskFulfillment.FulfillmentId
		data.FulfillmentStatus = confirmed.KioskFulfillment.Status
	}

	// The whole order and bill for the caller standing at the till: it is about to ask for money and
	// the amount is here rather than one request away.
	view, err := services.LoadConfirmedOrderView(ctx, confirmed.SalesOrderId, confirmed.InitialBillId)
	if err != nil {
		return nil, err
	}
	if view != nil {
		data.Order = view.Order
		data.InitialBill = view.Bill
	}

	return &it.ConfirmSalesOrderResult{HasData: true, Data: data}, nil
}

func (this *SalesOrderExtServiceImpl) CancelOrder(
	ctx corectx.Context, command it.CancelSalesOrderCommand,
) (*it.CancelSalesOrderResult, error) {
	if cErrs := assertPermission(
		ctx, composable.PermissionUpdate, c.SalesOrderResource, c.ResourceScopeOrg); cErrs != nil {
		return &it.CancelSalesOrderResult{ClientErrors: *cErrs}, nil
	}

	cancelled, vErrs, err := services.CancelOrder(
		ctx, command.SalesOrderId, command.Reason, this.dLock, this.reservations)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &it.CancelSalesOrderResult{ClientErrors: *vErrs}, nil
	}

	return &it.CancelSalesOrderResult{
		HasData: true,
		Data: it.CancelledOrderData{
			SalesOrderId:           cancelled.SalesOrderId,
			Status:                 cancelled.Status,
			ReleasedFulfillmentIds: cancelled.ReleasedFulfillmentIds,
		},
	}, nil
}

// salesFulfillmentMethodService resolves the derived method service through the resource hub.
//
// Read here rather than injected because this package is imported by dynamicengines, which builds
// the onions: injecting would close an import cycle. A resource that is not built answers nil, and
// resolution then treats the sale as an ordinary one.
func salesFulfillmentMethodService() *services.SalesFulfillmentMethodDomainServiceImpl {
	return services.FulfillmentMethodService()
}

func (this *SalesOrderExtServiceImpl) CreateAttempt(
	ctx corectx.Context, command it.CreateAttemptCommand,
) (*it.CreateAttemptResult, error) {
	// Update on the order, not a permission of its own: commanding a dispense changes what a
	// delivery owes, which is the same power over the same sale that recording its result is.
	if cErrs := assertPermission(
		ctx, composable.PermissionUpdate, c.SalesOrderResource, c.ResourceScopeOrg); cErrs != nil {
		return &it.CreateAttemptResult{ClientErrors: *cErrs}, nil
	}

	items := make([]services.CreateAttemptItem, 0, len(command.Items))
	for _, item := range command.Items {
		items = append(items, services.CreateAttemptItem{
			FulfillmentItemId: item.FulfillmentItemId,
			Quantity:          item.Quantity,
		})
	}

	created, vErrs, err := services.CreateAttempt(ctx, services.CreateAttemptParams{
		FulfillmentId:    command.FulfillmentId,
		ExecutorOutletId: command.ExecutorOutletId,
		Items:            items,
	}, this.dLock)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &it.CreateAttemptResult{ClientErrors: *vErrs}, nil
	}

	return &it.CreateAttemptResult{
		HasData: true,
		Data: it.AttemptData{
			AttemptId:             created.AttemptId,
			AttemptNo:             created.AttemptNo,
			ExternalCorrelationId: created.ExternalCorrelationId,
		},
	}, nil
}

func (this *SalesOrderExtServiceImpl) ReportAttemptResult(
	ctx corectx.Context, command it.ReportAttemptResultCommand,
) (*it.ReportAttemptResultResult, error) {
	if cErrs := assertPermission(
		ctx, composable.PermissionUpdate, c.SalesOrderResource, c.ResourceScopeOrg); cErrs != nil {
		return &it.ReportAttemptResultResult{ClientErrors: *cErrs}, nil
	}

	items := make([]services.ApplyAttemptResultItem, 0, len(command.Items))
	for _, item := range command.Items {
		items = append(items, services.ApplyAttemptResultItem{
			FulfillmentItemId: item.FulfillmentItemId,
			DispensedQty:      item.DispensedQty,
			FailedQty:         item.FailedQty,
			FailureCode:       item.FailureCode,
			FailureMessage:    item.FailureMessage,
		})
	}

	policy := services.ResolveSalesPolicy(ctx, this.settings)
	outcome, vErrs, err := services.ApplyAttemptResult(ctx, services.ApplyAttemptResultParams{
		AttemptId:             command.AttemptId,
		ResultEventId:         command.ResultEventId,
		ExternalCorrelationId: command.ExternalCorrelationId,
		InventoryResultRef:    command.InventoryResultRef,
		ExecutorOutletId:      command.ExecutorOutletId,
		Items:                 items,
	}, this.dLock, policy)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &it.ReportAttemptResultResult{ClientErrors: *vErrs}, nil
	}

	data := it.AttemptResultData{
		AttemptId:         outcome.AttemptId,
		AttemptStatus:     outcome.AttemptStatus,
		FulfillmentId:     outcome.FulfillmentId,
		FulfillmentStatus: outcome.FulfillmentStatus,
		AlreadyApplied:    outcome.AlreadyApplied,
	}
	if outcome.Refund != nil {
		data.RefundId = outcome.Refund.RefundId
		data.RefundStatus = outcome.Refund.RefundStatus
	}
	return &it.ReportAttemptResultResult{HasData: true, Data: data}, nil
}

func (this *SalesOrderExtServiceImpl) ViewFulfillment(
	ctx corectx.Context, command it.ViewFulfillmentCommand,
) (*it.ViewFulfillmentResult, error) {
	if cErrs := assertPermission(
		ctx, composable.PermissionRead, c.SalesOrderResource, c.ResourceScopeOrg); cErrs != nil {
		return &it.ViewFulfillmentResult{ClientErrors: *cErrs}, nil
	}

	view, err := services.ViewFulfillment(ctx, command.FulfillmentId)
	if err != nil {
		return nil, err
	}
	if view == nil {
		return &it.ViewFulfillmentResult{}, nil
	}

	items := make([]it.FulfillmentItemView, 0, len(view.Items))
	for _, item := range view.Items {
		items = append(items, it.FulfillmentItemView{
			FulfillmentItemId: item.ItemId,
			ProductVariantId:  item.ProductVariantId,
			RemainingQty:      item.RemainingQty,
			PendingRefundQty:  item.PendingRefundQty,
			FulfillableQty:    item.FulfillableQty,
			SourceLocationId:  item.SourceLocationId,
			InventorySourceId: item.InventorySourceId,
		})
	}

	return &it.ViewFulfillmentResult{
		HasData: true,
		Data: it.FulfillmentView{
			FulfillmentId:     view.FulfillmentId,
			FulfillmentStatus: view.Status,
			TargetOutletId:    view.TargetOutletId,
			Items:             items,
		},
	}, nil
}
