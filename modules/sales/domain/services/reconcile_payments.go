package services

import (
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// Reconciling payments whose verdict never arrived.
//
// THE ANNOUNCEMENT IS AN OPTIMIZATION; THIS IS THE GUARANTEE. The event bus acknowledges a message
// before a subscriber has handled it, so a crash in the wrong moment loses a settlement for good.
// Without this sweep a customer could pay and the bill stay open forever, which is the one failure
// mode a till cannot be asked to live with.
//
// It asks paymentinvoice what became of each order rather than guessing from elapsed time: only the
// gateway knows whether a customer paid, and a sweep that expired payments on a timer would refuse
// money that had in fact arrived.

// ReconcilePaymentsResult reports what one pass did.
type ReconcilePaymentsResult struct {
	Examined int
	Settled  int
	Failed   int

	// Unknown counts orders paymentinvoice has no record of. Worth surfacing separately: it means
	// Sales holds a correlation nothing issued, which no retry will fix.
	Unknown int
}

// ReconcileStalePayments applies the verdicts that were never announced.
//
// olderThan keeps the sweep off payments that were only just opened: the customer may still be
// standing at the terminal, and asking the gateway about every fresh order would be a round trip per
// payment for an answer that has not been decided yet.
func ReconcileStalePayments(
	ctx corectx.Context,
	orders itExt.PaymentOrderExtService,
	now time.Time,
	olderThan time.Duration,
	limit int,
) (*ReconcilePaymentsResult, error) {
	result := &ReconcilePaymentsResult{}
	if orders == nil {
		// No port, nothing to ask. Not an error: a build without the gateway still runs, and the
		// sweep simply has no work it can do.
		return result, nil
	}

	stale, err := pendingGatewayPaymentsOlderThan(ctx, now.Add(-olderThan), limit)
	if err != nil {
		return nil, err
	}

	for _, payment := range stale {
		result.Examined++

		orderId := stringOf(payment, models.SalesPaymentFieldPaymentOrderId)
		status, err := orders.GetOrderStatus(ctx, orderId)
		if err != nil {
			// One unreachable order must not end the pass: the rest are independent, and this one
			// comes round again on the next tick.
			continue
		}
		if !status.Found {
			result.Unknown++
			continue
		}

		outcome, decided := reconcileOutcomeOf(*status)
		if !decided {
			// Still being paid. Left alone deliberately — the customer has not finished.
			continue
		}

		applied, err := ConfirmPaymentAndSettle(ctx, ConfirmPaymentParams{
			PaymentOrderId:   orderId,
			SalesPaymentId:   stringOf(payment, models.SalesPaymentFieldId),
			Outcome:          outcome,
			RefTransactionId: status.RefTransactionId,
		})
		if err != nil {
			continue
		}
		if applied == nil || !applied.Applied {
			continue
		}

		if outcome == ConfirmPaymentPaid {
			result.Settled++
		} else {
			result.Failed++
		}
	}
	return result, nil
}

// reconcileOutcomeOf maps what the gateway says onto a verdict, or reports that there is none yet.
func reconcileOutcomeOf(status itExt.GatewayOrderStatus) (ConfirmPaymentOutcome, bool) {
	switch {
	case status.Settled:
		return ConfirmPaymentPaid, true
	case status.Failed:
		return ConfirmPaymentFailed, true
	}
	return "", false
}

// pendingGatewayPaymentsOlderThan finds the payments still waiting on a verdict.
//
// The age filter is pushed into the query rather than applied in Go. The comment on
// draftOrdersOlderThan claims RepoSearchParam carries no timestamp comparison; that is stale —
// dmodel.Operator has LessThan and GreaterThan — and reading every pending payment to discard most
// of them would grow with the table.
func pendingGatewayPaymentsOlderThan(
	ctx corectx.Context, cutoff time.Time, limit int,
) ([]dmodel.DynamicFields, error) {
	engine, err := engineFor(models.SalesPaymentSchemaName)
	if err != nil {
		return nil, err
	}

	size := limit
	if size <= 0 {
		size = model.MODEL_RULE_PAGE_MAX_SIZE
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			models.SalesPaymentFieldStatus, dmodel.Equals,
			string(models.SalesPaymentStatusPending)),
		*dmodel.NewSearchNode().NewCondition(
			basemodel.FieldCreatedAt, dmodel.LessThan, model.ModelDateTime(cutoff)),
	)

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  size,
	})
	if err != nil {
		return nil, err
	}
	if found == nil || !found.HasData {
		return nil, nil
	}

	// A cash payment left pending by a till has no order to ask about, so it is dropped here rather
	// than in the query: the column is null for those, and a null-aware predicate would differ
	// between the two migration trees for no gain.
	gateway := make([]dmodel.DynamicFields, 0, len(found.Data.Items))
	for _, payment := range found.Data.Items {
		if stringOf(payment, models.SalesPaymentFieldPaymentOrderId) != "" {
			gateway = append(gateway, payment)
		}
	}
	return gateway, nil
}
