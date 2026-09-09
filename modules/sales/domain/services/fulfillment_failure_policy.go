package services

import (
	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// What happens to goods a machine could not hand over.
//
// The policy is read from the fulfillment's SNAPSHOT, never from the method catalogue. A customer
// bought under a rule that said their money comes back automatically, and an operator editing the
// catalogue afterwards must not be able to take that away — nor to grant it to a sale that was made
// without it.
//
// Nothing here refunds delivered goods, and nothing refunds a whole order because one item failed.
// A shortfall is a quantity, and the remedy applies to exactly that quantity.

// FailurePolicyRefund describes a refund the policy asked for. Raising one is NOT the same as
// refunding: the quantity is not refunded until settlement says so, which is why this carries a
// status rather than a completion.
type FailurePolicyRefund struct {
	RefundId     string
	RefundStatus string

	// Quantity is what the refund was raised for — the terminally failed amount, never more.
	Quantity decimal.Decimal
}

// ApplyFailurePolicy decides what a fulfillment's outstanding quantity means now.
//
// It returns the status the fulfillment should take, or "" to leave the caller's own derivation
// alone. Only a TERMINAL failure changes anything: a shortfall that may still be retried is simply
// outstanding, and moving the fulfillment on would strand goods the customer could still receive.
func ApplyFailurePolicy(
	ctx corectx.Context,
	fulfillment dmodel.DynamicFields,
	items []dmodel.DynamicFields,
	attemptId string,
	policy SalesPolicy,
) (*FailurePolicyRefund, models.FulfillmentStatus, error) {
	outstanding := decimal.Zero
	for _, record := range items {
		outstanding = outstanding.Add(
			models.NewSalesOrderFulfillmentItemFrom(record).RemainingQuantity())
	}
	if !outstanding.IsPositive() {
		// Nothing is owed. Whatever failed on the way was either retried successfully or refunded,
		// and the completion rule the caller applies is the whole answer.
		return nil, "", nil
	}

	attempts, err := attemptsOfFulfillment(ctx,
		stringOf(fulfillment, models.SalesOrderFulfillmentFieldId))
	if err != nil {
		return nil, "", err
	}
	if !isTerminallyFailed(fulfillment, int32(len(attempts))) {
		// Retries remain. The shortfall stays outstanding and attemptable, which is exactly what a
		// customer with attempts left should get.
		return nil, "", nil
	}

	action := models.FulfillmentFailureAction(
		stringOf(fulfillment, models.SalesOrderFulfillmentFieldFailureAction))
	switch action {
	case models.FulfillmentFailureActionAutoRefund:
		// The money goes back with nobody asked, which is what an anonymous walk-up sale needs: there
		// is no customer to come back to. It runs through the SAME operation a person raising a
		// refund uses, so the two cannot drift into refunding differently.
		refund, err := raiseAutomaticRefund(ctx, fulfillment, items, attemptId, policy)
		if err != nil {
			return nil, "", err
		}
		// The fulfillment still waits rather than completing: a raised refund is not a paid one, and
		// the quantity stays owed until settlement says otherwise.
		return refund, models.FulfillmentStatusWaitingCustomerAction, nil

	case models.FulfillmentFailureActionCustomerActionRequired:
		// No refund, and deliberately so: the customer keeps the entitlement and decides what
		// happens next — another kiosk, another attempt, or their money back.
		return nil, models.FulfillmentStatusWaitingCustomerAction, nil

	case models.FulfillmentFailureActionManualResolution:
		// Parked for an operator. The status is the whole action: nothing is refunded and nothing is
		// retried automatically.
		return nil, models.FulfillmentStatusWaitingCustomerAction, nil
	}
	return nil, "", nil
}

// isTerminallyFailed reports that no further attempt will be made on its own.
//
// A NULL max_attempts is never terminal by counter. That is the point of the null: the customer
// decides when to stop rather than a number, and a fulfillment that expired its attempts by policy
// would take that decision away from them.
func isTerminallyFailed(fulfillment dmodel.DynamicFields, attemptCount int32) bool {
	maxAttempts := optionalInt32Of(fulfillment, models.SalesOrderFulfillmentFieldMaxAttempts)
	if maxAttempts == nil {
		return false
	}
	return attemptCount >= *maxAttempts
}

// TerminallyFailedQuantity is what a refund may be raised for once attempts are exhausted: the
// outstanding quantity, and never a unit more.
//
// Delivered goods are excluded because the customer has them, and quantity already refunded is
// excluded because it was paid back once — both are already out of RemainingQuantity, which is why
// this reads that rather than summing failures across attempts. A failure counted per attempt would
// grow with every retry and refund the same unit twice.
func TerminallyFailedQuantity(item dmodel.DynamicFields) decimal.Decimal {
	return models.NewSalesOrderFulfillmentItemFrom(item).RemainingQuantity()
}

// raiseAutomaticRefund creates the refund a terminal failure owes, for exactly the failed quantity.
//
// At-most-once, guarded by looking for a refund this fulfillment already raised. The guard matters
// because a result can be delivered twice: the idempotency gate upstream absorbs most replays, but a
// retry that legitimately re-enters the policy must find the existing refund rather than raise a
// second one for goods that failed once.
//
// Never refunds delivered goods, and never refunds the whole order because one item failed: the
// quantity is each item's own outstanding amount, and an item that was fully delivered contributes
// nothing.
func raiseAutomaticRefund(
	ctx corectx.Context,
	fulfillment dmodel.DynamicFields,
	items []dmodel.DynamicFields,
	attemptId string,
	policy SalesPolicy,
) (*FailurePolicyRefund, error) {
	fulfillmentId := stringOf(fulfillment, models.SalesOrderFulfillmentFieldId)

	existing, err := existingFulfillmentRefund(ctx, fulfillmentId)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	lines := make([]CreateReturnLine, 0, len(items))
	total := decimal.Zero
	for _, record := range items {
		owed := TerminallyFailedQuantity(record)
		if !owed.IsPositive() {
			continue
		}
		total = total.Add(owed)
		lines = append(lines, CreateReturnLine{
			SalesOrderLineId:  stringOf(record, models.SalesOrderFulfillmentItemFieldSalesOrderLineId),
			FulfillmentId:     fulfillmentId,
			FulfillmentItemId: stringOf(record, models.SalesOrderFulfillmentItemFieldId),

			// No goods are coming back — they never left the machine — so the goods quantity is zero
			// while the refund quantity is what was owed.
			Quantity:     decimal.Zero,
			RequestedQty: owed,
		})
	}
	if len(lines) == 0 {
		return nil, nil
	}

	orderId := stringOf(fulfillment, models.SalesOrderFulfillmentFieldSalesOrderId)
	order, err := loadRecord(ctx, models.SalesOrderSchemaName, models.SalesOrderFieldId, orderId)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, nil
	}

	// createReturnUnderLock rather than CreateReturn: the caller already holds this order's lock,
	// and the lock is a plain Redis SETNX with no re-entrancy — going through the public entry point
	// would spin against ourselves and then refuse every automatic refund.
	created, vErrs, err := createReturnUnderLock(ctx, CreateReturnParams{
		SalesOrderId: orderId,
		Reason:       "Automatic refund for goods a fulfillment could not hand over (attempt " + attemptId + ")",
		RefundReason: models.SalesRefundReasonFulfillmentFailure,
		ReturnType:   models.SalesReturnTypeRefundOnly,
		Lines:        lines,
	}, order, policy)
	if err != nil {
		return nil, err
	}
	if vErrs != nil && vErrs.Count() > 0 {
		// A refusal here is not the customer's problem and must not fail the result that caused it:
		// the goods genuinely did not come out, and that has already been recorded. The fulfillment
		// parks as waiting_customer_action, where an operator can see the shortfall and act.
		return nil, nil
	}

	return &FailurePolicyRefund{
		RefundId:     created.SalesReturnId,
		RefundStatus: created.RefundStatus,
		Quantity:     total,
	}, nil
}

// existingFulfillmentRefund finds a refund already raised against this fulfillment, so a second pass
// over the same terminal failure adopts it instead of paying the customer twice.
func existingFulfillmentRefund(
	ctx corectx.Context, fulfillmentId string,
) (*FailurePolicyRefund, error) {
	lines, err := searchBy(ctx, models.SalesReturnLineSchemaName,
		models.SalesReturnLineFieldFulfillmentId, fulfillmentId)
	if err != nil || len(lines) == 0 {
		return nil, err
	}

	returnId := stringOf(lines[0], models.SalesReturnLineFieldSalesReturnId)
	salesReturn, err := loadRecord(ctx, models.SalesReturnSchemaName,
		models.SalesReturnFieldId, returnId)
	if err != nil || salesReturn == nil {
		return nil, err
	}
	if !models.NewSalesReturnFrom(salesReturn).IsFulfillmentFailure() {
		// A customer-requested refund against the same delivery is somebody else's decision and must
		// not be mistaken for the automatic one; the automatic path may still raise its own.
		return nil, nil
	}

	quantity := decimal.Zero
	for _, line := range lines {
		quantity = quantity.Add(models.NewSalesReturnLineFrom(line).RefundQuantityRequested())
	}
	return &FailurePolicyRefund{
		RefundId:     returnId,
		RefundStatus: stringOf(salesReturn, models.SalesReturnFieldRefundStatus),
		Quantity:     quantity,
	}, nil
}
