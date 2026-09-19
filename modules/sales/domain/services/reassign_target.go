package services

import (
	"time"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	lock "github.com/sky-as-code/nikki-erp/modules/core/infra/distributedlock"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// Moving a delivery to another machine.
//
// This is a business operation, never a field update. Writing target_outlet_id directly would leave
// the goods held at the old location while the order promised them from the new one — the record and
// the shelf disagreeing, with the customer standing in front of the wrong machine. The stock has to
// move with the promise or neither moves.
//
// Only the OUTSTANDING quantity is reallocated. Goods already dispensed are the customer's and are
// not at the old kiosk to move; asking Inventory to shift them would claim stock for a delivery that
// already happened.
//
// Two paths, and the difference matters. A fulfillment that still holds its stock has that hold
// MOVED atomically — claimed at the destination before released at the origin, so a customer
// promised another kiosk cannot lose the goods to another buyer in the gap. A fulfillment whose
// reservation has already lapsed has nothing to move, so a fresh hold is taken instead; the history
// records that distinctly rather than implying an unbroken reservation that never existed.

const (
	ReasonTargetChangeNotAllowed    = "sales.fulfillment.target_change_not_allowed"
	ReasonTargetChangeTerminal      = "sales.fulfillment.terminal"
	ReasonTargetChangeNothingOwed   = "sales.fulfillment.nothing_outstanding"
	ReasonTargetChangeActiveAttempt = "sales.fulfillment.attempt_in_progress"
	ReasonTargetChangeSameOutlet    = "sales.fulfillment.already_at_target"
)

// ReassignTargetParams is one request to move a delivery.
type ReassignTargetParams struct {
	FulfillmentId string

	// ToOutletId is the sales point the goods should come from now.
	ToOutletId string

	// Reason is why, in business terms. Defaulted to customer_requested when unstated, which is the
	// overwhelmingly common case and the least presumptuous reading of a bare request.
	Reason models.TargetChangeReason
}

// ReassignTargetResult reports where the delivery now points and how it got there.
type ReassignTargetResult struct {
	FulfillmentId string
	FromOutletId  string
	ToOutletId    string
	Status        string

	// InventoryReference is the hold at the new location.
	InventoryReference string

	// ReservationExpiresAt is the fresh deadline, RFC3339, or empty when the policy sets no TTL.
	// A reassignment restarts the clock: the customer is being promised a new machine now, not
	// inheriting however little was left of the old promise.
	ReservationExpiresAt string

	// Reserved distinguishes a moved hold from a newly taken one, which is the difference between
	// "your goods followed you" and "your goods were re-found somewhere else".
	FreshReservation bool
}

// ReassignTarget moves a fulfillment to another sales point, stock and all.
func ReassignTarget(
	ctx corectx.Context,
	params ReassignTargetParams,
	dLock lock.DistributedLock,
	reservations itExt.FulfillmentReservationExtService,
) (*ReassignTargetResult, *ft.ClientErrors, error) {
	if dLock == nil {
		// Without the lock a reassignment can interleave with an attempt: the goods move while a
		// machine is dispensing them, and both believe they hold the reservation.
		return nil, nil, errors.New(
			"the distributed lock is not available; a fulfillment target cannot be reassigned without it")
	}

	fulfillment, err := loadRecord(ctx, models.SalesOrderFulfillmentSchemaName,
		models.SalesOrderFulfillmentFieldId, params.FulfillmentId)
	if err != nil {
		return nil, nil, err
	}
	if fulfillment == nil {
		return nil, targetChangeRefusal(ReasonFulfillmentNotFound,
			"no fulfillment with id "+params.FulfillmentId), nil
	}

	orderId := stringOf(fulfillment, models.SalesOrderFulfillmentFieldSalesOrderId)
	key := confirmLockKeyOf(orderId)
	acquired, err := dLock.AcquireWithRetry(
		ctx, key, confirmLockTtl, confirmLockRetryCount, confirmLockRetryDelay)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "acquiring the lock of order '%s'", orderId)
	}
	if !acquired {
		return nil, targetChangeRefusal(ReasonLockUnavailable,
			"this order is being changed by another request; try again"), nil
	}
	defer func() { _ = dLock.Release(ctx, key) }()

	// Re-read under the lock. Every guard below turns on the fulfillment's status and its
	// outstanding quantity, and a copy read while queuing may predate the attempt that just settled.
	fulfillment, err = loadRecord(ctx, models.SalesOrderFulfillmentSchemaName,
		models.SalesOrderFulfillmentFieldId, params.FulfillmentId)
	if err != nil {
		return nil, nil, err
	}
	if fulfillment == nil {
		return nil, targetChangeRefusal(ReasonFulfillmentNotFound,
			"no fulfillment with id "+params.FulfillmentId), nil
	}
	return reassignUnderLock(ctx, params, fulfillment, reservations)
}

func reassignUnderLock(
	ctx corectx.Context,
	params ReassignTargetParams,
	fulfillment dmodel.DynamicFields,
	reservations itExt.FulfillmentReservationExtService,
) (*ReassignTargetResult, *ft.ClientErrors, error) {
	if vErrs, err := assertOrderNotExpiredById(ctx, stringOf(fulfillment, models.SalesOrderFulfillmentFieldSalesOrderId)); err != nil || vErrs != nil {
		return nil, vErrs, err
	}
	orgId := stringOf(fulfillment, basemodel.FieldOrgId)
	fromOutletId := stringOf(fulfillment, models.SalesOrderFulfillmentFieldTargetOutletId)

	if vErrs := assertReassignable(fulfillment, params, fromOutletId); vErrs != nil {
		return nil, vErrs, nil
	}

	// The destination must be able to execute a fulfillment at all, checked exactly as it was when
	// the order was first confirmed — same org, enabled, and naming an inventory location.
	toLocationId, refusal, err := assertTargetUsable(ctx, params.ToOutletId, orgId)
	if err != nil {
		return nil, nil, err
	}
	if refusal != nil && refusal.ClientErrors.Count() > 0 {
		return nil, &refusal.ClientErrors, nil
	}

	items, err := ItemsOfFulfillment(ctx, params.FulfillmentId)
	if err != nil {
		return nil, nil, err
	}
	outstanding, reservationItems := outstandingReservationItems(items)
	if !outstanding.IsPositive() {
		// Nothing is owed, so there is nothing to move. Refused rather than treated as a no-op: a
		// caller asking to move a finished delivery has misunderstood something, and silently
		// succeeding would hide it.
		return nil, targetChangeRefusal(ReasonTargetChangeNothingOwed,
			"this fulfillment owes no goods, so there is nothing to move"), nil
	}

	vErrs, err := assertNoActiveAttempt(ctx, params.FulfillmentId)
	if err != nil || vErrs != nil {
		return nil, vErrs, err
	}

	record := models.NewSalesOrderFulfillmentFrom(fulfillment)
	expiresAt := reservationDeadline(
		optionalInt32Of(fulfillment, models.SalesOrderFulfillmentFieldReservationTtlMinutes),
		time.Now().UTC())

	// The lapsed case takes a FRESH hold rather than moving one: the old reservation is gone, its
	// goods are back in the sellable pool, and asking Inventory to reallocate something it no longer
	// holds would refuse for a reason that has nothing to do with the destination's stock.
	freshReservation := record.HasLapsed(time.Now().UTC()) ||
		models.FulfillmentStatus(stringOf(fulfillment,
			models.SalesOrderFulfillmentFieldFulfillmentStatus)) == models.FulfillmentStatusExpired

	response, vErrs, err := moveReservation(ctx, moveReservationParams{
		FulfillmentId: params.FulfillmentId,
		OrgId:         orgId,
		ToLocationId:  toLocationId,
		ExpiresAt:     expiresAt,
		Items:         reservationItems,
		Fresh:         freshReservation,
	}, reservations)
	if err != nil || vErrs != nil {
		// Inventory refused. The old target and its hold are untouched — that is the whole point of
		// claiming before releasing — so the customer keeps exactly the promise they had.
		return nil, vErrs, err
	}

	reason := params.Reason
	if reason == "" {
		reason = models.TargetChangeReasonCustomerRequested
	}
	if freshReservation {
		// Recorded distinctly: the stock genuinely went back to the pool in between, and the history
		// should not suggest an unbroken hold.
		reason = models.TargetChangeReasonReservationExpired
	}

	inventoryRef := ""
	if response != nil {
		inventoryRef = response.InventoryReference
	}
	status := models.FulfillmentStatusReserved
	err = commitReassignment(ctx, commitReassignmentParams{
		Fulfillment:  fulfillment,
		FromOutletId: fromOutletId,
		ToOutletId:   params.ToOutletId,
		OrgId:        orgId,
		InventoryRef: inventoryRef,
		Reason:       reason,
		ExpiresAt:    expiresAt,
		Status:       status,
		Items:        items,
	})
	if err != nil {
		return nil, nil, err
	}

	result := &ReassignTargetResult{
		FulfillmentId:      params.FulfillmentId,
		FromOutletId:       fromOutletId,
		ToOutletId:         params.ToOutletId,
		Status:             string(status),
		InventoryReference: inventoryRef,
		FreshReservation:   freshReservation,
	}
	if expiresAt != nil {
		result.ReservationExpiresAt = expiresAt.GoTime().Format(time.RFC3339)
	}
	return result, nil, nil
}

// assertReassignable covers the guards that read only the fulfillment and the request.
func assertReassignable(
	fulfillment dmodel.DynamicFields, params ReassignTargetParams, fromOutletId string,
) *ft.ClientErrors {
	record := models.NewSalesOrderFulfillmentFrom(fulfillment)

	if params.ToOutletId == "" {
		return targetChangeRefusal(ReasonTargetRequired,
			"name the sales point to move this fulfillment to")
	}
	if params.ToOutletId == fromOutletId {
		// Refused rather than treated as a no-op: moving stock to where it already is would take a
		// real reservation apart and rebuild it for no reason.
		return targetChangeRefusal(ReasonTargetChangeSameOutlet,
			"this fulfillment already delivers from that sales point")
	}

	// The POLICY the customer bought under decides this, not the current catalogue. A sale made at
	// the machine the customer was standing at promised no second machine, and an operator editing
	// the method afterwards must not be able to grant one.
	if !boolOf(fulfillment, models.SalesOrderFulfillmentFieldAllowTargetChange) {
		return targetChangeRefusal(ReasonTargetChangeNotAllowed,
			"the method this order was sold under does not allow its target to change")
	}
	if record.IsTerminal() {
		return targetChangeRefusal(ReasonTargetChangeTerminal,
			"a completed or cancelled fulfillment cannot be moved")
	}
	return nil
}

// assertNoActiveAttempt refuses while a machine may still be dispensing. Moving the goods out from
// under an outstanding attempt would let the old kiosk hand over stock that is now reserved
// somewhere else entirely.
//
// A read failure is returned as a Go error rather than swallowed: "could not find out" and "there
// are none" are different answers, and treating the first as the second would move the goods out
// from under an attempt nobody managed to look for.
func assertNoActiveAttempt(
	ctx corectx.Context, fulfillmentId string,
) (*ft.ClientErrors, error) {
	attempts, err := attemptsOfFulfillment(ctx, fulfillmentId)
	if err != nil {
		return nil, err
	}
	for _, attempt := range attempts {
		if models.NewSalesFulfillmentAttemptFrom(attempt).IsOutstanding() {
			return targetChangeRefusal(ReasonTargetChangeActiveAttempt,
				"an attempt on this fulfillment is still awaiting its result; the goods cannot "+
					"move while a machine may still be dispensing them"), nil
		}
	}
	return nil, nil
}

// outstandingReservationItems builds the demand to hold at the new location: what is still owed, and
// nothing else. Goods already dispensed are the customer's and are not at the old kiosk to move.
func outstandingReservationItems(
	items []dmodel.DynamicFields,
) (decimal.Decimal, []itExt.FulfillmentReservationItem) {
	total := decimal.Zero
	reservationItems := make([]itExt.FulfillmentReservationItem, 0, len(items))

	for _, record := range items {
		remaining := models.NewSalesOrderFulfillmentItemFrom(record).RemainingQuantity()
		if !remaining.IsPositive() {
			continue
		}
		total = total.Add(remaining)
		reservationItems = append(reservationItems, itExt.FulfillmentReservationItem{
			FulfillmentItemId: stringOf(record, models.SalesOrderFulfillmentItemFieldId),
			ProductVariantId:  stringOf(record, models.SalesOrderFulfillmentItemFieldProductVariantId),
			UomId:             stringOf(record, models.SalesOrderFulfillmentItemFieldUomId),
			Quantity:          remaining,
		})
	}
	return total, reservationItems
}

type moveReservationParams struct {
	FulfillmentId string
	OrgId         string
	ToLocationId  string
	ExpiresAt     *model.ModelDateTime
	Items         []itExt.FulfillmentReservationItem
	Fresh         bool
}

// moveReservation asks Inventory to hold the outstanding goods at the new location.
//
// A partial hold is refused for the same reason it is at confirm: the customer is being promised one
// machine, and half the goods there is a different promise they should be able to decline. Because
// Inventory claims the destination before releasing the origin, a refusal leaves the original hold
// exactly as it was.
func moveReservation(
	ctx corectx.Context,
	params moveReservationParams,
	reservations itExt.FulfillmentReservationExtService,
) (*itExt.FulfillmentReservationResponse, *ft.ClientErrors, error) {
	if reservations == nil {
		// No port bound. Refused rather than proceeding: moving the target while the stock stayed put
		// would promise goods from a machine that is not holding any.
		return nil, targetChangeRefusal(ReasonFulfillmentReserveFailed,
			"no inventory port is bound, so the stock cannot be moved with the target"), nil
	}

	request := itExt.FulfillmentReservationRequest{
		FulfillmentId: params.FulfillmentId,
		OrgId:         params.OrgId,
		LocationId:    params.ToLocationId,
		ExpiresAt:     params.ExpiresAt,
		Items:         params.Items,
	}

	var response *itExt.FulfillmentReservationResponse
	var err error
	if params.Fresh {
		response, err = reservations.ReserveForFulfillment(ctx, request)
	} else {
		response, err = reservations.ReallocateReservation(ctx, request)
	}
	if err != nil {
		return nil, nil, err
	}

	if response == nil || !response.Accepted {
		return nil, reservationRefusal(response, ReasonFulfillmentReserveFailed), nil
	}
	if !response.FullyReserved {
		return nil, reservationRefusal(response, ReasonFulfillmentInsufficientStock), nil
	}
	return response, nil, nil
}

type commitReassignmentParams struct {
	Fulfillment  dmodel.DynamicFields
	FromOutletId string
	ToOutletId   string
	OrgId        string
	InventoryRef string
	Reason       models.TargetChangeReason
	ExpiresAt    *model.ModelDateTime
	Status       models.FulfillmentStatus
	Items        []dmodel.DynamicFields
}

// commitReassignment records the change only after Inventory has actually moved the goods, in one
// transaction: a target updated without its audit row would lose the evidence of why a customer's
// order changed machine, and an audit row without the target would describe a move that did not
// happen.
func commitReassignment(ctx corectx.Context, params commitReassignmentParams) error {
	return withTransaction(ctx, models.SalesOrderFulfillmentSchemaName, func(tranxCtx corectx.Context) error {
		changes := dmodel.DynamicFields{
			models.SalesOrderFulfillmentFieldTargetOutletId:    params.ToOutletId,
			models.SalesOrderFulfillmentFieldFulfillmentStatus: string(params.Status),
		}
		if params.ExpiresAt != nil {
			changes[models.SalesOrderFulfillmentFieldReservationExpiresAt] = *params.ExpiresAt
		} else {
			// A method with no TTL must not inherit the old deadline, or the fresh promise would
			// expire on the previous one's clock.
			changes[models.SalesOrderFulfillmentFieldReservationExpiresAt] = nil
		}
		err := writeChanges(tranxCtx,
			models.SalesOrderFulfillmentSchemaName, params.Fulfillment, changes)
		if err != nil {
			return err
		}

		if err := repointItemReservations(tranxCtx, params.Items, params.InventoryRef); err != nil {
			return err
		}
		return writeTargetChange(tranxCtx, params)
	})
}

// repointItemReservations moves every item's inventory reference to the new hold, so an operator
// tracing one item is not shown a reservation that was released.
func repointItemReservations(
	ctx corectx.Context, items []dmodel.DynamicFields, inventoryRef string,
) error {
	for _, record := range items {
		if models.NewSalesOrderFulfillmentItemFrom(record).IsSettled() {
			// A settled item's goods already left; its old reference is history and stays as it is.
			continue
		}
		err := writeChanges(ctx, models.SalesOrderFulfillmentItemSchemaName, record,
			dmodel.DynamicFields{
				models.SalesOrderFulfillmentItemFieldInventoryReservationRef: inventoryRef,
			})
		if err != nil {
			return err
		}
	}
	return nil
}

func writeTargetChange(ctx corectx.Context, params commitReassignmentParams) error {
	engineRepo, err := repoFor(models.SalesFulfillmentTargetChangeSchemaName)
	if err != nil {
		return err
	}
	id, err := model.NewId()
	if err != nil {
		return err
	}

	actorType, actorId := describeActor(ctx)
	fields := dmodel.DynamicFields{
		models.SalesFulfillmentTargetChangeFieldId:            string(*id),
		models.SalesFulfillmentTargetChangeFieldFulfillmentId: stringOf(params.Fulfillment, models.SalesOrderFulfillmentFieldId),
		models.SalesFulfillmentTargetChangeFieldToOutletId:    params.ToOutletId,
		models.SalesFulfillmentTargetChangeFieldReason:        string(params.Reason),
		models.SalesFulfillmentTargetChangeFieldActorType:     string(actorType),
		models.SalesFulfillmentTargetChangeFieldChangedAt:     model.ModelDateTime(time.Now().UTC()),
		basemodel.FieldOrgId:                                  params.OrgId,
	}
	if params.FromOutletId != "" {
		fields[models.SalesFulfillmentTargetChangeFieldFromOutletId] = params.FromOutletId
	}
	if params.InventoryRef != "" {
		fields[models.SalesFulfillmentTargetChangeFieldInventoryOperationRef] = params.InventoryRef
	}
	if actorId != "" {
		fields[models.SalesFulfillmentTargetChangeFieldActorId] = actorId
	}

	_, err = engineRepo.Insert(ctx, fields)
	return errors.Wrap(err, "writing the fulfillment target change")
}

// describeActor reads who is acting from the request's own authentication, never from the payload.
// A caller able to name its own actor could attribute its changes to somebody else.
//
// An unauthenticated execution is `system`: that is what a scheduled sweep is, and inventing a user
// for it would put a person's name against a change no person made.
func describeActor(ctx corectx.Context) (models.TargetChangeActorType, string) {
	if ctx == nil {
		return models.TargetChangeActorTypeSystem, ""
	}
	principal := ctx.GetPermissions().Principal
	if principal.IsZero() {
		return models.TargetChangeActorTypeSystem, ""
	}

	switch principal.Kind {
	case corectx.PrincipalKindUser:
		return models.TargetChangeActorTypeUser, string(principal.Id)
	case corectx.PrincipalKindService:
		return models.TargetChangeActorTypeService, string(principal.Id)
	}
	return models.TargetChangeActorTypeSystem, string(principal.Id)
}

// TargetChangesOfFulfillment lists a delivery's move history, oldest first as stored.
func TargetChangesOfFulfillment(
	ctx corectx.Context, fulfillmentId string,
) ([]dmodel.DynamicFields, error) {
	return searchBy(ctx, models.SalesFulfillmentTargetChangeSchemaName,
		models.SalesFulfillmentTargetChangeFieldFulfillmentId, fulfillmentId)
}

func targetChangeRefusal(reason, message string) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(
		models.SalesOrderFulfillmentSchemaName, reason, message))
	return vErrs
}
