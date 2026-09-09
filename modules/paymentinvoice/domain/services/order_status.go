package services

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
	itOrder "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/order"
)

// Reading an order back.
//
// Settlement is announced rather than returned: an order goes out to a gateway and its verdict
// arrives later, through a callback or the watchdog. Every announcement can be missed — the result
// sync can exhaust its retries, and the event bus acknowledges a message before a subscriber has
// handled it — so a caller whose payment has been pending too long needs somewhere to ask.
//
// This is that place, and it is the reason the calling module can treat announcements as an
// optimization rather than a guarantee.

type (
	GetOrderStatusQuery      = itOrder.GetOrderStatusQuery
	GetOrderStatusResultData = itOrder.GetOrderStatusResultData
	GetOrderStatusResult     = itOrder.GetOrderStatusResult
)

// GetOrderStatus answers where an order stands.
//
// A missing order is a rule the caller broke, not a failure: it means they hold an identifier this
// module never issued, which they can act on. It travels in ClientErrors like every other refusal
// on this port.
func (this *OrderDomainService) GetOrderStatus(
	ctx corectx.Context, query GetOrderStatusQuery,
) (*GetOrderStatusResult, error) {
	vErrs := ft.NewClientErrors()

	order, err := findOrderToRead(ctx, query)
	if err != nil {
		return nil, err
	}
	if order == nil {
		appendFieldViolation(vErrs, models.OrderFieldOrderId,
			"paymentinvoice.order_not_found", "no order with "+describeReadTarget(query))
		return &GetOrderStatusResult{ClientErrors: *vErrs}, nil
	}

	// The gateway's own identifier lives on the completed payment transaction, not the order, so it
	// is read through the same helper a refund uses to find what it is reversing. Empty while the
	// order is unsettled, which is the answer a caller reconciling a pending payment expects.
	refTransactionId, err := findPaymentRefTransactionId(ctx, derefString(order.GetId()))
	if err != nil {
		return nil, err
	}

	return &GetOrderStatusResult{
		HasData: true,
		Data: GetOrderStatusResultData{
			OrderId:          derefString(order.GetOrderId()),
			OrderCode:        derefString(order.GetOrderCode()),
			Status:           derefString(order.GetStatus()),
			Amount:           derefDecimal(order.GetAmount()),
			RefundAmount:     derefDecimal(order.GetRefundAmount()),
			RefTransactionId: refTransactionId,
		},
	}, nil
}

// findOrderToRead accepts either identifier, for the same reason a refund does: the standalone
// service handed both out together and callers stored whichever they happened to keep.
func findOrderToRead(ctx corectx.Context, query GetOrderStatusQuery) (*models.Order, error) {
	if query.OrderId != "" {
		return findOrderByBusinessId(ctx, query.OrderId)
	}
	if query.OrderCode != "" {
		return findOrderByCode(ctx, query.OrderCode)
	}
	return nil, nil
}

// describeReadTarget names what the caller quoted, so "not found" says which identifier missed.
func describeReadTarget(query GetOrderStatusQuery) string {
	if query.OrderId != "" {
		return "id '" + query.OrderId + "'"
	}
	if query.OrderCode != "" {
		return "code '" + query.OrderCode + "'"
	}
	return "no identifier at all"
}
