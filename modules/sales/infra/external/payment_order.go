package external

import (
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itOrder "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/order"

	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// An adapter rather than a direct hand-over, for two reasons that both keep paymentinvoice's shapes
// out of Sales' domain.
//
// The conventions this integration runs on live here and nowhere else: the source tag stamped on
// every order Sales opens, the metadata a settlement is matched back on, and the fact that Sales
// deliberately passes no return_url because the verdict comes back in-process rather than over HTTP.
// Scattering those through the domain would mean a second caller could get any of them wrong.
//
// And paymentinvoice's order states are mapped to the three answers reconciliation acts on, so its
// state machine does not end up duplicated in Sales where it would go stale.

// salesOrderSource tags every order Sales opens, and becomes the leading characters of the order id
// paymentinvoice hands out. Frozen once live: support reads it to tell where an order came from, and
// the vending machines' own tag is the precedent.
const salesOrderSource = "SALES"

type paymentOrderAdapter struct {
	orders itOrder.OrderDomainService
}

func (this *paymentOrderAdapter) CreatePayment(
	ctx corectx.Context, cmd itExt.CreateGatewayPaymentCommand,
) (*itExt.CreateGatewayPaymentResult, error) {
	content := cmd.Content
	var contentPtr *string
	if content != "" {
		contentPtr = &content
	}

	created, err := this.orders.CreatePayment(ctx, itOrder.CreatePaymentCommand{
		OrgId:           cmd.OrgId,
		Amount:          cmd.Amount,
		PaymentMethodId: cmd.PaymentMethodId,
		Source:          salesOrderSource,
		Content:         contentPtr,

		// No ReturnUrl: that field asks paymentinvoice to POST the outcome back over HTTP, which is
		// how the vending machines are told. Sales is in the same process and hears the settlement
		// event instead, so asking for a callback would mean the same verdict arriving twice.
		ReturnUrl: nil,

		// The correlation, carried by the order itself. sales_payment.payment_order_id is the key a
		// settlement is normally matched on; this is the copy that survives if that column was never
		// written — the order can be opened and the process die before the id comes back.
		Metadata: map[string]any{
			"sales_payment_id": cmd.SalesPaymentId,
			"sales_bill_id":    cmd.SalesBillId,
			"org_id":           cmd.OrgId,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "opening a payment order")
	}
	if created == nil {
		return nil, errors.New("opening a payment order returned nothing")
	}

	if created.ClientErrors.Count() > 0 {
		return &itExt.CreateGatewayPaymentResult{
			Refused:       true,
			RefusalReason: firstRefusal(&created.ClientErrors),
		}, nil
	}
	if !created.HasData {
		return &itExt.CreateGatewayPaymentResult{
			Refused:       true,
			RefusalReason: "the payment provider did not open an order",
		}, nil
	}

	return &itExt.CreateGatewayPaymentResult{
		HasData: true,
		Data: itExt.CreateGatewayPaymentResultData{
			OrderId:   created.Data.OrderId,
			OrderCode: created.Data.OrderCode,
			QrCodeUrl: created.Data.QrCodeUrl,
			PayUrl:    created.Data.PayUrl,
		},
	}, nil
}

func (this *paymentOrderAdapter) Refund(
	ctx corectx.Context, cmd itExt.RefundGatewayPaymentCommand,
) (*itExt.RefundGatewayPaymentResult, error) {
	content := cmd.Content
	var contentPtr *string
	if content != "" {
		contentPtr = &content
	}

	refunded, err := this.orders.Refund(ctx, itOrder.RefundCommand{
		OrderId: cmd.OrderId,
		Amount:  cmd.Amount,
		Content: contentPtr,
	})
	if err != nil {
		return nil, errors.Wrap(err, "refunding a payment order")
	}
	if refunded == nil {
		return nil, errors.New("refunding a payment order returned nothing")
	}

	if refunded.ClientErrors.Count() > 0 {
		return &itExt.RefundGatewayPaymentResult{
			Refused:       true,
			RefusalReason: firstRefusal(&refunded.ClientErrors),
		}, nil
	}
	if !refunded.HasData {
		return &itExt.RefundGatewayPaymentResult{
			Refused:       true,
			RefusalReason: "the payment provider did not confirm the refund",
		}, nil
	}

	return &itExt.RefundGatewayPaymentResult{
		HasData: true,
		Data: itExt.RefundGatewayPaymentResultData{
			OrderId:      refunded.Data.OrderId,
			RefundAmount: refunded.Data.RefundAmount,
			// "Rested" is the old service's spelling, kept upstream because callers read that key.
			// Sales names it for what it is; the rename stops here rather than spreading.
			RemainingAmount: refunded.Data.RestedAmount,
		},
	}, nil
}

func (this *paymentOrderAdapter) GetOrderStatus(
	ctx corectx.Context, orderId string,
) (*itExt.GatewayOrderStatus, error) {
	found, err := this.orders.GetOrderStatus(ctx, itOrder.GetOrderStatusQuery{OrderId: orderId})
	if err != nil {
		return nil, errors.Wrapf(err, "reading payment order '%s'", orderId)
	}
	if found == nil || !found.HasData {
		// Not found is an answer, not a failure: it means Sales holds an order id this module never
		// issued, which reconciliation reports rather than retries forever.
		return &itExt.GatewayOrderStatus{Found: false}, nil
	}

	return &itExt.GatewayOrderStatus{
		Found:            true,
		OrderId:          found.Data.OrderId,
		Settled:          isSettledOrderStatus(found.Data.Status),
		Failed:           isFailedOrderStatus(found.Data.Status),
		RefTransactionId: found.Data.RefTransactionId,
		Amount:           found.Data.Amount,
		RefundAmount:     found.Data.RefundAmount,
	}, nil
}

// isSettledOrderStatus reports whether the money is in.
//
// A refunded order counts as settled: it was paid, and what happened afterwards is a refund Sales
// tracks on its own refund legs. Reconciliation asking "did this payment ever complete" must not be
// told no simply because the money was later given back.
func isSettledOrderStatus(status string) bool {
	switch status {
	case itOrder.OrderStatusPaymentSuccess,
		itOrder.OrderStatusRefundSuccess,
		itOrder.OrderStatusRefundFailed:
		return true
	}
	return false
}

// isFailedOrderStatus reports whether the collection is over without the money arriving. Pending and
// processing are neither settled nor failed — the customer may still be paying.
func isFailedOrderStatus(status string) bool {
	switch status {
	case itOrder.OrderStatusPaymentFailed,
		itOrder.OrderStatusCanceled,
		itOrder.OrderStatusExpired:
		return true
	}
	return false
}

// firstRefusal flattens the upstream's refusal into the one line a till can show. The full set stays
// upstream: Sales reports why a payment could not start, not paymentinvoice's whole rule book.
func firstRefusal(vErrs *ft.ClientErrors) string {
	if vErrs == nil || vErrs.Count() == 0 {
		return "the payment was refused"
	}
	for _, violation := range *vErrs {
		if message := violation.String(); message != "" {
			return message
		}
	}
	return "the payment was refused"
}
