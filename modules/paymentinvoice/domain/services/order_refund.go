package services

import (
	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
	itGateway "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/gateway"
)

// RefundCommand asks for money to be given back against an order the caller quotes by its
// business identifier — the one they were given, not this module's primary key.
type RefundCommand struct {
	OrderId string
	Amount  decimal.Decimal
	Content *string
}

// RefundResult reports what was returned and what remains.
type RefundResult struct {
	OrderId      string
	RefundAmount decimal.Decimal

	// RestedAmount is what is left of the order after this refund. The spelling is the old
	// service's and is kept: the ordering system reads this key.
	RestedAmount decimal.Decimal
}

// Refund gives money back for an order already paid.
//
// The guard rails run in a fixed order and every one of them is a client error, because each
// describes something about the caller's request rather than a failure here. They are checked
// before the gateway is called: asking a gateway to reverse a payment that was never collected
// gets an opaque refusal, and the caller learns more from being told which rule they broke.
//
// Two corrections to the old service's rules are deliberate:
//
//   - It refused a second refund after a failed one (refund_failed was treated as terminal),
//     which left an order stuck when the gateway had been briefly unavailable. Only a successful
//     refund closes the order here.
//   - It compared each refund against the order total in isolation, so an order could be refunded
//     repeatedly for the full amount. The running total is checked instead.
func (this *OrderDomainService) Refund(
	ctx corectx.Context, cmd RefundCommand,
) (*RefundResult, *ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()

	order, err := findOrderByBusinessId(ctx, cmd.OrderId)
	if err != nil {
		return nil, vErrs, err
	}
	if order == nil {
		appendFieldViolation(vErrs, models.OrderFieldOrderId,
			"paymentinvoice.order_not_found", "no order with id '"+cmd.OrderId+"'")
		return nil, vErrs, nil
	}

	if !assertRefundable(*order, cmd.Amount, vErrs) {
		return nil, vErrs, nil
	}

	method, err := this.loadRefundMethod(ctx, *order, vErrs)
	if err != nil || vErrs.Count() > 0 {
		return nil, vErrs, err
	}

	adapter, exists := this.registry.Get(derefString(method.GetAdapterCode()))
	if !exists {
		appendOrderViolation(vErrs, "paymentinvoice.gateway_unavailable",
			"payment method '"+derefString(method.GetCode())+"' is not available on this deployment")
		return nil, vErrs, nil
	}

	assertAmountWithinMethodBounds(cmd.Amount, method, vErrs)
	if vErrs.Count() > 0 {
		return nil, vErrs, nil
	}

	return this.reverse(ctx, adapter, *order, *method, cmd, vErrs)
}

// reverse calls the gateway and records the outcome against the order.
func (this *OrderDomainService) reverse(
	ctx corectx.Context,
	adapter itGateway.PaymentGateway,
	order models.Order,
	method models.PaymentMethod,
	cmd RefundCommand,
	vErrs *ft.ClientErrors,
) (*RefundResult, *ft.ClientErrors, error) {
	paymentRefId, err := findPaymentRefTransactionId(ctx, derefString(order.GetId()))
	if err != nil {
		return nil, vErrs, err
	}

	// The refund goes back through the account the payment came in on. Any other account would
	// have the gateway refuse it — it has no such transaction — and if it did not, the money would
	// leave the wrong merchant's balance.
	profileConfig, err := this.profileConfigForOrder(ctx, order)
	if err != nil {
		return nil, vErrs, err
	}

	refunded, gatewayErr := adapter.Refund(ctx, itGateway.RefundRequest{
		OrderCode:        derefString(order.GetOrderCode()),
		Amount:           cmd.Amount,
		Content:          cmd.Content,
		RefTransactionId: paymentRefId,
		Metadata:         order.GetMetadata(),
		MethodConfig:     method.GetConfig(),
		ProfileConfig:    profileConfig,
	})

	if gatewayErr != nil {
		if err := this.markRefundFailed(ctx, order, method, cmd, gatewayErr); err != nil {
			return nil, vErrs, err
		}
		appendOrderViolation(vErrs, "paymentinvoice.refund_failed",
			"the payment gateway refused the refund: "+gatewayErr.Error())
		return nil, vErrs, nil
	}

	total := derefDecimal(order.GetRefundAmount()).Add(cmd.Amount)
	if err := this.markRefundSucceeded(ctx, order, method, cmd, refunded, total); err != nil {
		return nil, vErrs, err
	}

	return &RefundResult{
		OrderId:      derefString(order.GetOrderId()),
		RefundAmount: cmd.Amount,
		RestedAmount: derefDecimal(order.GetAmount()).Sub(total),
	}, vErrs, nil
}

// markRefundSucceeded closes the order and appends the refund transaction.
//
// The running total is written rather than the amount of this refund, so a partially refunded
// order still says how much of it has been given back.
func (this *OrderDomainService) markRefundSucceeded(
	ctx corectx.Context,
	order models.Order,
	method models.PaymentMethod,
	cmd RefundCommand,
	refunded *itGateway.RefundResult,
	total decimal.Decimal,
) error {
	metadata := mergeMetadata(order.GetMetadata(), map[string]any{
		models.OrderMetaRefundResponse: refunded.RawResponse,
	})

	return withOrderTransaction(ctx, func(tranxCtx corectx.Context) error {
		if err := writeOrderFields(tranxCtx, derefString(order.GetId()), dmodel.DynamicFields{
			models.OrderFieldStatus:       models.OrderStatusRefundSuccess,
			models.OrderFieldRefundAmount: total,
			models.OrderFieldMetadata:     metadata,
		}); err != nil {
			return err
		}
		return appendRefundTransaction(tranxCtx, order, method, cmd,
			models.TransactionStatusCompleted, refunded.RefTransactionId, refunded.RawResponse)
	})
}

// markRefundFailed records the attempt without closing the order, so it can be retried.
func (this *OrderDomainService) markRefundFailed(
	ctx corectx.Context,
	order models.Order,
	method models.PaymentMethod,
	cmd RefundCommand,
	gatewayErr error,
) error {
	payload := map[string]any{"error": gatewayErr.Error()}
	metadata := mergeMetadata(order.GetMetadata(), map[string]any{
		models.OrderMetaRefundResponse: payload,
	})

	return withOrderTransaction(ctx, func(tranxCtx corectx.Context) error {
		if err := writeOrderFields(tranxCtx, derefString(order.GetId()), dmodel.DynamicFields{
			models.OrderFieldStatus:   models.OrderStatusRefundFailed,
			models.OrderFieldMetadata: metadata,
		}); err != nil {
			return err
		}
		return appendRefundTransaction(tranxCtx, order, method, cmd,
			models.TransactionStatusFailed, "", payload)
	})
}

// appendRefundTransaction records one refund attempt against the order.
//
// A row is written whether the attempt succeeded or failed, because the transactions are the
// evidence of what was tried, not only of what worked.
func appendRefundTransaction(
	ctx corectx.Context,
	order models.Order,
	method models.PaymentMethod,
	cmd RefundCommand,
	status string,
	refTransactionId string,
	payload map[string]any,
) error {
	fields := dmodel.DynamicFields{
		models.TransactionFieldOrderId:         derefString(order.GetId()),
		models.TransactionFieldOrderBusinessId: derefString(order.GetOrderId()),
		models.TransactionFieldStatus:          status,
		models.TransactionFieldAmount:          cmd.Amount,
		models.TransactionFieldCurrencyId:      derefString(order.GetCurrencyId()),
		models.TransactionFieldPaymentMethodId: derefString(method.GetId()),
		models.TransactionFieldTransactionType: models.TransactionTypeRefund,
	}
	if cmd.Content != nil && *cmd.Content != "" {
		fields[models.TransactionFieldContent] = *cmd.Content
	}
	if refTransactionId != "" {
		fields[models.TransactionFieldRefTransactionId] = refTransactionId
	}
	if payload != nil {
		fields[models.TransactionFieldRefPayload] = payload
	}

	_, err := createRecord(ctx, models.TransactionSchemaName, fields)
	return err
}

// assertRefundable applies the guard rails, in the order they are meant to be read.
//
// It returns false as soon as one fails, so the caller is told the first thing wrong with their
// request rather than a list in which the later entries are consequences of the first.
func assertRefundable(order models.Order, amount decimal.Decimal, vErrs *ft.ClientErrors) bool {
	status := derefString(order.GetStatus())

	if status == models.OrderStatusRefundSuccess {
		appendOrderViolation(vErrs, "paymentinvoice.order_already_refunded",
			"this order has already been refunded")
		return false
	}
	if status == models.OrderStatusCanceled || status == models.OrderStatusExpired {
		appendOrderViolation(vErrs, "paymentinvoice.order_not_payable",
			"a "+status+" order collected no money, so there is nothing to give back")
		return false
	}
	// refund_failed is deliberately allowed through: a previous attempt that the gateway
	// refused is a reason to try again, not a reason the order is closed.
	if status != models.OrderStatusPaymentSuccess && status != models.OrderStatusRefundFailed {
		appendOrderViolation(vErrs, "paymentinvoice.order_not_paid",
			"only a paid order can be refunded; this one is "+status)
		return false
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		appendFieldViolation(vErrs, models.OrderFieldAmount,
			"paymentinvoice.amount_not_positive", "the refund amount must be greater than zero")
		return false
	}

	total := derefDecimal(order.GetRefundAmount()).Add(amount)
	if total.GreaterThan(derefDecimal(order.GetAmount())) {
		appendFieldViolation(vErrs, models.OrderFieldAmount,
			"paymentinvoice.refund_exceeds_order",
			"refunding "+amount.String()+" would return more than the order collected")
		return false
	}
	return true
}

// loadRefundMethod fetches the method the order was paid by.
//
// The method is read even when it has since been withdrawn from use: switching a method off stops
// new payments, and refusing to give money back for payments already taken through it would be a
// different and much worse thing. This is why it does not go through loadActiveMethod.
func (this *OrderDomainService) loadRefundMethod(
	ctx corectx.Context, order models.Order, vErrs *ft.ClientErrors,
) (*models.PaymentMethod, error) {
	method, err := this.readMethod(ctx, derefString(order.GetPaymentMethodId()))
	if err != nil {
		return nil, err
	}
	if method == nil {
		appendOrderViolation(vErrs, "paymentinvoice.payment_method_not_found",
			"the payment method this order was paid by no longer exists")
		return nil, nil
	}
	return method, nil
}

// findPaymentRefTransactionId returns the gateway's identifier for the payment being reversed.
//
// A refund is filed against the original payment, and every gateway here identifies that payment
// by its own reference rather than by our order code. An empty result is left to the adapter to
// refuse: whether the reference is required is the gateway's rule, not this module's.
func findPaymentRefTransactionId(ctx corectx.Context, orderPk string) (string, error) {
	engine, err := engineFor(models.TransactionSchemaName)
	if err != nil {
		return "", err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			models.TransactionFieldOrderId, dmodel.Equals, orderPk),
		*dmodel.NewSearchNode().NewCondition(
			models.TransactionFieldTransactionType, dmodel.Equals, models.TransactionTypePayment),
		*dmodel.NewSearchNode().NewCondition(
			models.TransactionFieldStatus, dmodel.Equals, models.TransactionStatusCompleted),
	)

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  1,
	})
	if err != nil {
		return "", err
	}
	if found == nil || !found.HasData || len(found.Data.Items) == 0 {
		return "", nil
	}
	return derefString(models.NewTransactionFrom(found.Data.Items[0]).GetRefTransactionId()), nil
}
