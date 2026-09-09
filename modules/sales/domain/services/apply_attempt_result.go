package services

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	lock "github.com/sky-as-code/nikki-erp/modules/core/infra/distributedlock"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// Recording what an executor actually did, and settling everything that follows from it.
//
// This is the one operation in the feature that must be safe to deliver twice. A dispense result may
// arrive over a message bus that is at-most-once and over REST as a backstop, so the same physical
// event legitimately reaches here more than once. Replaying it must not add the delivered quantity
// again, write the attempt items twice, or raise a second refund — so the event id is stored, and a
// redelivery is answered from the record rather than re-applied.
//
// A redelivery is distinguished from a CONTRADICTION by hashing the payload: the same id with the
// same content is the network doing its job, while the same id with different content is two
// different claims about one physical event, which is refused rather than guessed at.
//
// Sales acts only AFTER Inventory has applied the physical consequence, which is why an
// inventory_result_ref is required. A sale that recorded a dispense Inventory had not yet consumed
// would report goods gone while the stock ledger still believes they are on the shelf.

const (
	ReasonResultEventConflict    = "sales.fulfillment.result_event_conflict"
	ReasonResultQuantityMismatch = "sales.fulfillment.result_quantity_mismatch"
	ReasonResultEventRequired    = "sales.fulfillment.result_event_id_required"
	ReasonResultInventoryRef     = "sales.fulfillment.inventory_result_ref_required"
	ReasonResultCorrelation      = "sales.fulfillment.correlation_mismatch"
	ReasonResultAlreadyReported  = "sales.fulfillment.attempt_already_reported"
	ReasonResultItemUnknown      = "sales.fulfillment.result_item_not_in_attempt"
)

// ApplyAttemptResultParams is one executor's report about one try.
type ApplyAttemptResultParams struct {
	AttemptId string

	// ResultEventId is the REPORTER's id for this event, and the idempotency key. Required: without
	// one, a redelivery is indistinguishable from a second dispense.
	ResultEventId string

	// ExternalCorrelationId echoes what create-attempt handed out, so a reply that drifted onto the
	// wrong attempt is refused instead of settling somebody else's sale.
	ExternalCorrelationId string

	// InventoryResultRef proves Inventory already applied the physical consequence.
	InventoryResultRef string

	ExecutorOutletId string
	Items            []ApplyAttemptResultItem
}

// ApplyAttemptResultItem is what happened to one product.
type ApplyAttemptResultItem struct {
	FulfillmentItemId string
	DispensedQty      decimal.Decimal
	FailedQty         decimal.Decimal
	FailureCode       string
	FailureMessage    string
}

// ApplyAttemptResultOutcome is what the caller is told, and what a replay repeats verbatim.
type ApplyAttemptResultOutcome struct {
	AttemptId         string
	AttemptStatus     string
	FulfillmentId     string
	FulfillmentStatus string

	// AlreadyApplied marks the replay path: the answer is the stored one and nothing was written
	// again. A caller retrying after a lost reply gets success, not a conflict.
	AlreadyApplied bool

	// Refund is populated once the refund lifecycle exists and the failure policy raised one.
	// Creation is not success: the quantity is not refunded until settlement says so.
	Refund *FailurePolicyRefund
}

// ApplyAttemptResult records a result and settles the quantities, statuses and failure policy.
func ApplyAttemptResult(
	ctx corectx.Context,
	params ApplyAttemptResultParams,
	dLock lock.DistributedLock,
	policy SalesPolicy,
) (*ApplyAttemptResultOutcome, *ft.ClientErrors, error) {
	if dLock == nil {
		return nil, nil, errors.New(
			"the distributed lock is not available; an attempt result cannot be applied without it")
	}
	if strings.TrimSpace(params.ResultEventId) == "" {
		return nil, resultRefusal(ReasonResultEventRequired,
			"a result must carry an event id, which is what makes a redelivery safe"), nil
	}
	if strings.TrimSpace(params.InventoryResultRef) == "" {
		// Inventory owns the physical consequence. Recording a dispense it has not applied would
		// report goods gone while the stock ledger still holds them.
		return nil, resultRefusal(ReasonResultInventoryRef,
			"apply the result in Inventory first and pass its reference"), nil
	}

	attempt, err := loadRecord(ctx, models.SalesFulfillmentAttemptSchemaName,
		models.SalesFulfillmentAttemptFieldId, params.AttemptId)
	if err != nil {
		return nil, nil, err
	}
	if attempt == nil {
		return nil, resultRefusal(ReasonAttemptNotFound,
			"no fulfillment attempt with id "+params.AttemptId), nil
	}

	fulfillmentId := stringOf(attempt, models.SalesFulfillmentAttemptFieldFulfillmentId)
	fulfillment, err := loadRecord(ctx, models.SalesOrderFulfillmentSchemaName,
		models.SalesOrderFulfillmentFieldId, fulfillmentId)
	if err != nil {
		return nil, nil, err
	}
	if fulfillment == nil {
		return nil, resultRefusal(ReasonFulfillmentNotFound,
			"the attempt names a fulfillment that no longer exists"), nil
	}

	orderId := stringOf(fulfillment, models.SalesOrderFulfillmentFieldSalesOrderId)
	key := confirmLockKeyOf(orderId)
	acquired, err := dLock.AcquireWithRetry(
		ctx, key, confirmLockTtl, confirmLockRetryCount, confirmLockRetryDelay)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "acquiring the lock of order '%s'", orderId)
	}
	if !acquired {
		return nil, resultRefusal(ReasonLockUnavailable,
			"this order is being changed by another request; try again"), nil
	}
	defer func() { _ = dLock.Release(ctx, key) }()

	// Re-read under the lock. The idempotency decision turns on what has already been stored, and a
	// copy read while queuing may predate a concurrent delivery of this very event.
	attempt, err = loadRecord(ctx, models.SalesFulfillmentAttemptSchemaName,
		models.SalesFulfillmentAttemptFieldId, params.AttemptId)
	if err != nil {
		return nil, nil, err
	}
	if attempt == nil {
		return nil, resultRefusal(ReasonAttemptNotFound,
			"no fulfillment attempt with id "+params.AttemptId), nil
	}
	return applyResultUnderLock(ctx, params, attempt, orderId, policy)
}

func applyResultUnderLock(
	ctx corectx.Context,
	params ApplyAttemptResultParams,
	attempt dmodel.DynamicFields,
	orderId string,
	policy SalesPolicy,
) (*ApplyAttemptResultOutcome, *ft.ClientErrors, error) {
	fulfillmentId := stringOf(attempt, models.SalesFulfillmentAttemptFieldFulfillmentId)
	payloadHash := hashResultPayload(params)

	// The idempotency gate, before anything is written.
	if replay, vErrs := assertResultIsNew(attempt, params.ResultEventId, payloadHash); vErrs != nil {
		return nil, vErrs, nil
	} else if replay {
		return replayedOutcome(ctx, attempt, fulfillmentId)
	}

	if vErrs := assertCorrelationMatches(attempt, params); vErrs != nil {
		return nil, vErrs, nil
	}

	attemptItems, err := itemsOfAttempt(ctx, params.AttemptId)
	if err != nil {
		return nil, nil, err
	}
	reported, vErrs := matchReportedItems(attemptItems, params.Items)
	if vErrs != nil {
		return nil, vErrs, nil
	}

	if err := writeAttemptOutcome(ctx, params, attempt, reported, payloadHash); err != nil {
		return nil, nil, err
	}

	// Quantities settle before statuses, because every status below is derived from them.
	if err := SyncFulfillmentItemQuantities(ctx, fulfillmentId); err != nil {
		return nil, nil, err
	}
	// SFL-059: the fulfillment's quantities are the source, but cancel-vs-return gating, e-invoice
	// eligibility and return validation all read the ORDER. Without this the goods are delivered and
	// the order still says nothing was.
	if err := SyncOrderFulfillmentRollup(ctx, orderId); err != nil {
		return nil, nil, err
	}

	outcome, err := settleFulfillmentAfterResult(ctx, fulfillmentId, params.AttemptId, policy)
	if err != nil {
		return nil, nil, err
	}
	return outcome, nil, nil
}

// assertResultIsNew decides whether this delivery is new, a replay, or a contradiction.
//
// Returns (true, nil) for a replay the caller should answer from the record, (false, nil) for a new
// result to apply, and a refusal when the same event id arrives carrying different content — two
// claims about one physical event, which nothing here can adjudicate.
func assertResultIsNew(
	attempt dmodel.DynamicFields, eventId, payloadHash string,
) (bool, *ft.ClientErrors) {
	storedEventId := stringOf(attempt, models.SalesFulfillmentAttemptFieldResultEventId)
	if storedEventId == "" {
		if !models.NewSalesFulfillmentAttemptFrom(attempt).IsOutstanding() {
			// Settled without an event id: a cancelled attempt, which cannot then be reported on.
			return false, resultRefusal(ReasonResultAlreadyReported,
				"this attempt is no longer awaiting a result")
		}
		return false, nil
	}

	if storedEventId != eventId {
		// A different event against an attempt that already reported. The first result is the one
		// that happened; a second would double-count the goods.
		return false, resultRefusal(ReasonResultAlreadyReported,
			"this attempt has already reported its result")
	}
	if stringOf(attempt, models.SalesFulfillmentAttemptFieldResultPayloadHash) != payloadHash {
		return false, resultRefusal(ReasonResultEventConflict,
			"this event id was already applied with different content; one physical event cannot "+
				"have two outcomes")
	}
	return true, nil
}

// assertCorrelationMatches checks the reply belongs to this attempt and this machine. A reply that
// drifted from another attempt would settle a sale it never touched.
func assertCorrelationMatches(
	attempt dmodel.DynamicFields, params ApplyAttemptResultParams,
) *ft.ClientErrors {
	expected := stringOf(attempt, models.SalesFulfillmentAttemptFieldExternalCorrelationId)
	if params.ExternalCorrelationId != "" && expected != "" &&
		params.ExternalCorrelationId != expected {
		return resultRefusal(ReasonResultCorrelation,
			"this result does not carry the correlation id of the attempt it names")
	}

	executor := stringOf(attempt, models.SalesFulfillmentAttemptFieldExecutorOutletId)
	if params.ExecutorOutletId != "" && executor != "" && params.ExecutorOutletId != executor {
		return resultRefusal(ReasonAttemptWrongExecutor,
			"this result comes from a sales point other than the one that was asked")
	}
	return nil
}

// reportedItem pairs a stored attempt item with what the executor said about it.
type reportedItem struct {
	Record dmodel.DynamicFields
	Result ApplyAttemptResultItem
}

// matchReportedItems checks the report describes exactly this attempt, and that each line balances.
//
// Every requested item must be answered and no others. A partial report would leave items `pending`
// on a settled attempt, and an extra item would credit a sale for goods nobody asked for.
func matchReportedItems(
	attemptItems []dmodel.DynamicFields, reported []ApplyAttemptResultItem,
) ([]reportedItem, *ft.ClientErrors) {
	byFulfillmentItem := make(map[string]dmodel.DynamicFields, len(attemptItems))
	for _, record := range attemptItems {
		byFulfillmentItem[stringOf(record, models.SalesFulfillmentAttemptItemFieldFulfillmentItemId)] = record
	}

	matched := make([]reportedItem, 0, len(reported))
	seen := make(map[string]struct{}, len(reported))
	for _, item := range reported {
		record, known := byFulfillmentItem[item.FulfillmentItemId]
		if !known {
			return nil, resultRefusal(ReasonResultItemUnknown,
				"item "+item.FulfillmentItemId+" was not part of this attempt")
		}
		if _, repeated := seen[item.FulfillmentItemId]; repeated {
			// Two entries for one item could disagree, and nothing here could say which was right.
			return nil, resultRefusal(ReasonResultItemUnknown,
				"item "+item.FulfillmentItemId+" is reported twice in one result")
		}
		seen[item.FulfillmentItemId] = struct{}{}

		if item.DispensedQty.IsNegative() || item.FailedQty.IsNegative() {
			return nil, resultRefusal(ReasonResultQuantityMismatch,
				"item "+item.FulfillmentItemId+" reports a negative quantity")
		}

		attempted := decimalOf(record, models.SalesFulfillmentAttemptItemFieldAttemptedQty)
		if !item.DispensedQty.Add(item.FailedQty).Equal(attempted) {
			// Quantity that vanished, or appeared from nowhere. Refused rather than reconciled:
			// nothing here can say which of the three numbers is the wrong one.
			return nil, resultRefusal(ReasonResultQuantityMismatch,
				"item "+item.FulfillmentItemId+" must report dispensed plus failed equal to what "+
					"was attempted")
		}
		matched = append(matched, reportedItem{Record: record, Result: item})
	}

	if len(matched) != len(attemptItems) {
		return nil, resultRefusal(ReasonResultQuantityMismatch,
			"a result must report every item the attempt asked for")
	}
	return matched, nil
}

// writeAttemptOutcome stores the reported quantities and stamps the attempt, in one transaction so a
// half-applied result cannot exist — items updated without the event id recorded would be re-applied
// by the next redelivery and double-count the goods.
func writeAttemptOutcome(
	ctx corectx.Context,
	params ApplyAttemptResultParams,
	attempt dmodel.DynamicFields,
	reported []reportedItem,
	payloadHash string,
) error {
	return withTransaction(ctx, models.SalesFulfillmentAttemptSchemaName, func(tranxCtx corectx.Context) error {
		results := make([]models.FulfillmentAttemptItemResult, 0, len(reported))
		for _, item := range reported {
			result := models.DeriveItemResult(item.Result.DispensedQty, item.Result.FailedQty)
			results = append(results, result)

			changes := dmodel.DynamicFields{
				models.SalesFulfillmentAttemptItemFieldDispensedQty: item.Result.DispensedQty,
				models.SalesFulfillmentAttemptItemFieldFailedQty:    item.Result.FailedQty,
				models.SalesFulfillmentAttemptItemFieldItemResult:   string(result),
			}
			if item.Result.FailureCode != "" {
				changes[models.SalesFulfillmentAttemptItemFieldFailureCode] = item.Result.FailureCode
			}
			if item.Result.FailureMessage != "" {
				changes[models.SalesFulfillmentAttemptItemFieldFailureMessage] = item.Result.FailureMessage
			}
			err := writeChanges(tranxCtx,
				models.SalesFulfillmentAttemptItemSchemaName, item.Record, changes)
			if err != nil {
				return err
			}
		}

		completedAt := model.ModelDateTime(time.Now().UTC())
		return writeChanges(tranxCtx, models.SalesFulfillmentAttemptSchemaName, attempt,
			dmodel.DynamicFields{
				models.SalesFulfillmentAttemptFieldAttemptStatus:      string(models.DeriveAttemptStatus(results)),
				models.SalesFulfillmentAttemptFieldResultEventId:      params.ResultEventId,
				models.SalesFulfillmentAttemptFieldResultPayloadHash:  payloadHash,
				models.SalesFulfillmentAttemptFieldInventoryResultRef: params.InventoryResultRef,
				models.SalesFulfillmentAttemptFieldCompletedAt:        completedAt,
			})
	})
}

// SyncFulfillmentItemQuantities recomputes each fulfillment item's delivered quantity as the SUM of
// what its attempts dispensed.
//
// Recomputed, never incremented — the same reasoning as SyncFulfilledQuantities beside it. An
// increment compounds any update that did not land and double-counts any result that arrives twice,
// while a recount is self-correcting and makes a redelivery harmless even if the idempotency gate
// were somehow bypassed.
//
// Failed quantity is deliberately absent: goods that did not come out are still owed, and letting a
// failure reduce what is outstanding is the exact mistake the whole feature exists to prevent.
func SyncFulfillmentItemQuantities(ctx corectx.Context, fulfillmentId string) error {
	attempts, err := attemptsOfFulfillment(ctx, fulfillmentId)
	if err != nil {
		return err
	}

	dispensed := map[string]decimal.Decimal{}
	for _, attempt := range attempts {
		items, err := itemsOfAttempt(ctx, stringOf(attempt, models.SalesFulfillmentAttemptFieldId))
		if err != nil {
			return err
		}
		for _, item := range items {
			itemId := stringOf(item, models.SalesFulfillmentAttemptItemFieldFulfillmentItemId)
			dispensed[itemId] = dispensed[itemId].Add(
				decimalOf(item, models.SalesFulfillmentAttemptItemFieldDispensedQty))
		}
	}

	fulfillmentItems, err := ItemsOfFulfillment(ctx, fulfillmentId)
	if err != nil {
		return err
	}
	for _, record := range fulfillmentItems {
		itemId := stringOf(record, models.SalesOrderFulfillmentItemFieldId)
		total := dispensed[itemId]

		item := models.NewSalesOrderFulfillmentItemFrom(record)
		changes := dmodel.DynamicFields{
			models.SalesOrderFulfillmentItemFieldFulfilledQty: total,
			models.SalesOrderFulfillmentItemFieldItemStatus:   string(deriveItemStatus(*item, total)),
		}
		if err := writeChanges(ctx, models.SalesOrderFulfillmentItemSchemaName, record, changes); err != nil {
			return err
		}
	}
	return nil
}

// deriveItemStatus reads an item's own state off its quantities. Settled counts a refund as fully as
// a delivery: the customer is owed nothing either way.
func deriveItemStatus(
	item models.SalesOrderFulfillmentItem, fulfilled decimal.Decimal,
) models.FulfillmentItemStatus {
	ordered := decimal.Zero
	if value := item.GetOrderedQty(); value != nil {
		ordered = *value
	}
	refunded := decimal.Zero
	if value := item.GetRefundedQty(); value != nil {
		refunded = *value
	}

	switch {
	case fulfilled.Add(refunded).GreaterThanOrEqual(ordered):
		return models.FulfillmentItemStatusFulfilled
	case fulfilled.IsPositive():
		return models.FulfillmentItemStatusPartiallyFulfilled
	}
	return models.FulfillmentItemStatusReserved
}

// settleFulfillmentAfterResult moves the fulfillment to whatever its quantities now say, then lets
// the failure policy decide what happens to anything still outstanding.
func settleFulfillmentAfterResult(
	ctx corectx.Context, fulfillmentId, attemptId string, policy SalesPolicy,
) (*ApplyAttemptResultOutcome, error) {
	items, err := ItemsOfFulfillment(ctx, fulfillmentId)
	if err != nil {
		return nil, err
	}

	outstanding := decimal.Zero
	delivered := decimal.Zero
	for _, record := range items {
		item := models.NewSalesOrderFulfillmentItemFrom(record)
		outstanding = outstanding.Add(item.RemainingQuantity())
		if value := item.GetFulfilledQty(); value != nil {
			delivered = delivered.Add(*value)
		}
	}

	fulfillment, err := loadRecord(ctx, models.SalesOrderFulfillmentSchemaName,
		models.SalesOrderFulfillmentFieldId, fulfillmentId)
	if err != nil {
		return nil, err
	}
	if fulfillment == nil {
		return nil, errors.New("the fulfillment being settled no longer exists: " + fulfillmentId)
	}

	// Completion turns on outstanding quantity ALONE, whatever the refunds did on the way.
	status := models.FulfillmentStatusPartiallyFulfilled
	if outstanding.IsZero() {
		status = models.FulfillmentStatusCompleted
	} else if delivered.IsZero() {
		// Nothing came out at all. The fulfillment goes back to a resting state rather than staying
		// in_progress, which would block every further attempt on a delivery that is still owed.
		status = models.FulfillmentStatusReady
	}

	refund, policyStatus, err := ApplyFailurePolicy(ctx, fulfillment, items, attemptId, policy)
	if err != nil {
		return nil, err
	}
	if policyStatus != "" {
		status = policyStatus
	}

	if err := setFulfillmentStatus(ctx, fulfillmentId, status); err != nil {
		return nil, err
	}

	attempt, err := loadRecord(ctx, models.SalesFulfillmentAttemptSchemaName,
		models.SalesFulfillmentAttemptFieldId, attemptId)
	if err != nil {
		return nil, err
	}
	return &ApplyAttemptResultOutcome{
		AttemptId:         attemptId,
		AttemptStatus:     stringOf(attempt, models.SalesFulfillmentAttemptFieldAttemptStatus),
		FulfillmentId:     fulfillmentId,
		FulfillmentStatus: string(status),
		Refund:            refund,
	}, nil
}

// replayedOutcome answers a redelivery from what was stored, writing nothing. A caller retrying
// after a lost reply gets the original success rather than a conflict, which is what stops it
// retrying forever.
func replayedOutcome(
	ctx corectx.Context, attempt dmodel.DynamicFields, fulfillmentId string,
) (*ApplyAttemptResultOutcome, *ft.ClientErrors, error) {
	fulfillment, err := loadRecord(ctx, models.SalesOrderFulfillmentSchemaName,
		models.SalesOrderFulfillmentFieldId, fulfillmentId)
	if err != nil {
		return nil, nil, err
	}
	return &ApplyAttemptResultOutcome{
		AttemptId:         stringOf(attempt, models.SalesFulfillmentAttemptFieldId),
		AttemptStatus:     stringOf(attempt, models.SalesFulfillmentAttemptFieldAttemptStatus),
		FulfillmentId:     fulfillmentId,
		FulfillmentStatus: stringOf(fulfillment, models.SalesOrderFulfillmentFieldFulfillmentStatus),
		AlreadyApplied:    true,
	}, nil, nil
}

// hashResultPayload digests the parts of a report that assert what happened, so a redelivery can be
// told from a contradiction.
//
// Items are sorted before hashing: two deliveries of one event may legitimately order them
// differently, and treating that as a different payload would refuse a replay that agrees in every
// respect that matters. The failure text is included because a report claiming a different reason is
// a different claim, even at the same quantities.
func hashResultPayload(params ApplyAttemptResultParams) string {
	lines := make([]string, 0, len(params.Items))
	for _, item := range params.Items {
		lines = append(lines, strings.Join([]string{
			item.FulfillmentItemId,
			item.DispensedQty.String(),
			item.FailedQty.String(),
			item.FailureCode,
			item.FailureMessage,
		}, "\x1f"))
	}
	sort.Strings(lines)

	digest := sha256.Sum256([]byte(strings.Join(
		append([]string{params.InventoryResultRef}, lines...), "\x1e")))
	return hex.EncodeToString(digest[:])
}

func resultRefusal(reason, message string) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(
		models.SalesFulfillmentAttemptSchemaName, reason, message))
	return vErrs
}
