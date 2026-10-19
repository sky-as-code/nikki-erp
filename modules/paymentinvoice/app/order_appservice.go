package app

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/constants"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/services"
	itOrder "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/order"
)

// The order's authorized surface: the built-in CRUD plus the three actions that move money.
//
// Each action asserts its own permission code rather than "update". Being allowed to correct an
// order's description is not the same authority as being allowed to hand money back, and one
// permission covering both would make the smaller grant imply the larger. The actions are
// collection-level (none addresses an order by its record id), so only AssertAction runs.

type OrderApplicationServiceImpl struct {
	composable.CrudApplicationService
	orders *services.OrderDomainService
}

func NewOrderApplicationService(
	base composable.CrudApplicationService, orders *services.OrderDomainService,
) itOrder.OrderCrudApplicationService {
	return &OrderApplicationServiceImpl{CrudApplicationService: base, orders: orders}
}

// CreatePaymentAction records an order and asks its gateway to start collecting.
func (this *OrderApplicationServiceImpl) CreatePaymentAction(
	ctx corectx.Context, cmd itOrder.OrderActionCommand,
) (*itOrder.OrderActionResult, error) {
	if _, cErrs := this.AssertAction(ctx, constants.ActionCreatePayment, cmd); cErrs != nil {
		return actionFailure(cErrs), nil
	}
	payment, vErrs := buildCreatePaymentCommand(cmd)
	if vErrs.Count() > 0 {
		return actionFailure(vErrs), nil
	}
	result, err := this.orders.CreatePayment(ctx, payment)
	if err != nil {
		return nil, err
	}
	if result.ClientErrors.Count() > 0 {
		return actionFailure(&result.ClientErrors), nil
	}
	return actionSuccess(map[string]any{
		"order_id": result.Data.OrderId,
		// order_code is what the gateway knows the order by and what its callback arrives under,
		// so a caller reconciling against the gateway's own records needs it.
		"order_code":  result.Data.OrderCode,
		"qr_code_url": result.Data.QrCodeUrl,
		"pay_url":     result.Data.PayUrl,
	}), nil
}

// RefundAction gives money back for an order already paid.
func (this *OrderApplicationServiceImpl) RefundAction(
	ctx corectx.Context, cmd itOrder.OrderActionCommand,
) (*itOrder.OrderActionResult, error) {
	if _, cErrs := this.AssertAction(ctx, constants.ActionRefund, cmd); cErrs != nil {
		return actionFailure(cErrs), nil
	}
	refund, vErrs := buildRefundCommand(cmd)
	if vErrs.Count() > 0 {
		return actionFailure(vErrs), nil
	}
	result, err := this.orders.Refund(ctx, refund)
	if err != nil {
		return nil, err
	}
	if result.ClientErrors.Count() > 0 {
		return actionFailure(&result.ClientErrors), nil
	}
	// rested_amount keeps the old service's spelling: the ordering system reads this key.
	return actionSuccess(map[string]any{
		"order_id":      result.Data.OrderId,
		"refund_amount": result.Data.RefundAmount.String(),
		"rested_amount": result.Data.RestedAmount.String(),
	}), nil
}

// RemovePosOrdersAction clears the payment prompts queued on one card terminal.
func (this *OrderApplicationServiceImpl) RemovePosOrdersAction(
	ctx corectx.Context, cmd itOrder.OrderActionCommand,
) (*itOrder.OrderActionResult, error) {
	if _, cErrs := this.AssertAction(ctx, constants.ActionRemovePosOrders, cmd); cErrs != nil {
		return actionFailure(cErrs), nil
	}
	result, cErrs, err := this.orders.RemovePosOrders(ctx, services.RemovePosOrdersCommand{
		PosId: readString(cmd, paramPosId),
	})
	if err != nil {
		return nil, err
	}
	if cErrs.Count() > 0 {
		return actionFailure(cErrs), nil
	}
	return actionSuccess(map[string]any{"affected_count": result.AffectedCount}), nil
}

func actionFailure(cErrs *ft.ClientErrors) *itOrder.OrderActionResult {
	return &itOrder.OrderActionResult{ClientErrors: *cErrs}
}

func actionSuccess(data map[string]any) *itOrder.OrderActionResult {
	return &itOrder.OrderActionResult{HasData: true, Data: data}
}
