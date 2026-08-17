package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
)

// GatewayResult is what a gateway callback says became of one payment.
type GatewayResult struct {
	// OrderCode is the key the callback arrives under. It is the gateway's handle on the order,
	// never the order's primary key or its quoted order_id.
	OrderCode string

	// Paid is the verdict. False means the gateway reached an outcome and it was not payment;
	// a callback that reports nothing conclusive must not be turned into one of these.
	Paid bool

	// RefTransactionId is the gateway's own identifier for the settled payment, which is what a
	// later refund is filed against.
	RefTransactionId string

	// RefPayload is the callback body, kept verbatim as the evidence for this outcome.
	RefPayload map[string]any
}

// GatewayResultOutcome reports what applying a result did, so the caller can answer its gateway.
type GatewayResultOutcome struct {
	// Applied is false when nothing changed — either the order is unknown, or it had already
	// reached a terminal state.
	Applied bool

	// OrderFound distinguishes "no such order" from "already settled". VietQR needs the
	// distinction: it answers 001 for the first and success for the second.
	OrderFound bool

	// OrderId is the order's quoted identifier, which the result sync is keyed by.
	OrderId string
}

// ApplyGatewayResult records a gateway's verdict against an order and its payment transaction.
//
// All three callbacks converge here, because the state machine is the module's and not any one
// gateway's: what differs between them is how a request is authenticated and how a reply is
// shaped, and both of those stay in the transport layer.
//
// It is idempotent, and that is the point rather than a nicety. Every one of these gateways
// retries a callback it did not get a clean answer to, so the same result arriving twice is an
// ordinary event. A replay against an order that has already settled changes nothing and reports
// success: answering an error would have the gateway retry again, indefinitely.
//
// The re-check happens inside the transaction, so two callbacks racing cannot both apply — the
// second reads the state the first committed.
func (this *OrderDomainService) ApplyGatewayResult(
	ctx corectx.Context, result GatewayResult,
) (*GatewayResultOutcome, error) {
	if result.OrderCode == "" {
		return &GatewayResultOutcome{}, nil
	}

	outcome := &GatewayResultOutcome{}
	err := withOrderTransaction(ctx, func(tranxCtx corectx.Context) error {
		order, err := findOrderByCode(tranxCtx, result.OrderCode)
		if err != nil {
			return err
		}
		if order == nil {
			return nil
		}

		outcome.OrderFound = true
		outcome.OrderId = derefString(order.GetOrderId())

		// An order that has already reached a verdict keeps it. A late "failed" arriving after a
		// success would otherwise un-pay an order the customer has paid, and the goods have
		// likely already been released.
		if isTerminalPaymentStatus(derefString(order.GetStatus())) {
			return nil
		}

		if err := this.applyToOrder(tranxCtx, *order, result); err != nil {
			return err
		}
		outcome.Applied = true
		return nil
	})

	return outcome, err
}

// applyToOrder writes the verdict to the order and to its pending payment transaction.
func (this *OrderDomainService) applyToOrder(
	ctx corectx.Context, order models.Order, result GatewayResult,
) error {
	orderStatus := models.OrderStatusPaymentFailed
	transactionStatus := models.TransactionStatusFailed
	if result.Paid {
		orderStatus = models.OrderStatusPaymentSuccess
		transactionStatus = models.TransactionStatusCompleted
	}

	if err := writeOrderFields(ctx, derefString(order.GetId()), dmodel.DynamicFields{
		models.OrderFieldStatus: orderStatus,
	}); err != nil {
		return err
	}

	transaction, err := findPendingPaymentTransaction(ctx, derefString(order.GetId()))
	if err != nil {
		return err
	}
	if transaction == nil {
		// The order exists but its payment transaction does not, which the create flow makes
		// impossible: both are written in one transaction. The order's own state is already
		// correct, so this is recorded rather than failed — refusing the callback would have the
		// gateway retry against a state that will never change.
		return nil
	}

	fields := dmodel.DynamicFields{
		models.TransactionFieldStatus: transactionStatus,
	}
	if result.RefTransactionId != "" {
		fields[models.TransactionFieldRefTransactionId] = result.RefTransactionId
	}
	if result.RefPayload != nil {
		fields[models.TransactionFieldRefPayload] = result.RefPayload
	}
	return writeTransactionFields(ctx, derefString(transaction.GetId()), fields)
}

// isTerminalPaymentStatus reports whether an order has already reached a verdict.
//
// The refund states count as terminal here: an order that has been refunded was paid, and a
// payment callback replayed afterwards must not walk it back to payment_success.
func isTerminalPaymentStatus(status string) bool {
	switch status {
	case models.OrderStatusPaymentSuccess,
		models.OrderStatusPaymentFailed,
		models.OrderStatusCanceled,
		models.OrderStatusExpired,
		models.OrderStatusRefundSuccess,
		models.OrderStatusRefundFailed:
		return true
	}
	return false
}

// findOrderByCode looks an order up by the key its gateway callbacks arrive under.
func findOrderByCode(ctx corectx.Context, orderCode string) (*models.Order, error) {
	engine, err := engineFor(models.OrderSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.OrderFieldOrderCode, dmodel.Equals, orderCode),
	)

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  1,
	})
	if err != nil {
		return nil, errors.Wrap(err, "findOrderByCode")
	}
	if found == nil || !found.HasData || len(found.Data.Items) == 0 {
		return nil, nil
	}
	return models.NewOrderFrom(found.Data.Items[0]), nil
}

// findPendingPaymentTransaction returns the payment attempt awaiting a verdict.
//
// It filters on pending deliberately: an attempt already completed or failed has had its verdict,
// and a replayed callback must not overwrite the evidence recorded the first time.
func findPendingPaymentTransaction(
	ctx corectx.Context, orderPk string,
) (*models.Transaction, error) {
	engine, err := engineFor(models.TransactionSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			models.TransactionFieldOrderId, dmodel.Equals, orderPk),
		*dmodel.NewSearchNode().NewCondition(
			models.TransactionFieldTransactionType, dmodel.Equals, models.TransactionTypePayment),
		*dmodel.NewSearchNode().NewCondition(
			models.TransactionFieldStatus, dmodel.Equals, models.TransactionStatusPending),
	)

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  1,
	})
	if err != nil {
		return nil, errors.Wrap(err, "findPendingPaymentTransaction")
	}
	if found == nil || !found.HasData || len(found.Data.Items) == 0 {
		return nil, nil
	}
	return models.NewTransactionFrom(found.Data.Items[0]), nil
}
