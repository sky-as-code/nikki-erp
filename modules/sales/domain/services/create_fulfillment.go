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
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// Creating a kiosk fulfillment and holding its stock, as part of confirming an order.
//
// The ordering here is the opposite of the goods-issue path beside it, and deliberately so. That
// path confirms first and asks Inventory afterwards, because an unfulfillable order is a warehouse
// problem somebody can work through. A kiosk sale has nobody to work it through: the customer taps,
// pays, and expects a can to fall. So the stock is claimed BEFORE the money is taken, and a
// reservation that cannot be met refuses the confirm with the order still a draft — recoverable,
// re-confirmable at another kiosk, and with nothing charged.
//
// Everything in this file is gated on fulfillment_type = kiosk_dispense. No other channel's confirm
// changes, which is what keeps a reordering this significant safe to ship.

const (
	ReasonFulfillmentInsufficientStock = "sales.fulfillment.target_insufficient_stock"
	ReasonFulfillmentReserveFailed     = "sales.fulfillment.reservation_failed"
	ReasonFulfillmentNothingToDeliver  = "sales.fulfillment.nothing_to_deliver"
)

// CreateFulfillmentResult is what confirm reports about the delivery half of the sale.
type CreateFulfillmentResult struct {
	FulfillmentId  string
	Status         string
	TargetOutletId string

	// ReservationExpiresAt is when the hold lapses, RFC3339, or empty when the policy sets no TTL.
	ReservationExpiresAt string
}

// CreateKioskFulfillment resolves the policy, writes the fulfillment and its items, and holds the
// stock at the target.
//
// A nil result with no violations means this order is not a kiosk sale and nothing was done — the
// caller carries on with the untouched confirm path. That "not applicable" answer is deliberately
// distinct from a refusal: only one of them should stop a confirm.
func CreateKioskFulfillment(
	ctx corectx.Context,
	order dmodel.DynamicFields,
	methods *SalesFulfillmentMethodDomainServiceImpl,
	reservations itExt.FulfillmentReservationExtService,
) (*CreateFulfillmentResult, *ft.ClientErrors, error) {
	orderId := stringOf(order, models.SalesOrderFieldId)
	orgId := stringOf(order, basemodel.FieldOrgId)

	resolved, refusal, err := ResolveFulfillmentMethod(ctx, FulfillmentMethodRequest{
		FulfillmentMethodId:  stringOf(order, models.SalesOrderFieldRequestedFulfillmentMethodId),
		TargetOutletId:       stringOf(order, models.SalesOrderFieldRequestedTargetOutletId),
		SalesChannelId:       stringOf(order, models.SalesOrderFieldSalesChannelId),
		SalesPointId:         stringOf(order, models.SalesOrderFieldSalesPointId),
		OrgId:                orgId,
		CustomerIdentityMode: models.CustomerIdentityMode(stringOf(order, models.SalesOrderFieldCustomerIdentityMode)),
	}, methods)
	if err != nil {
		return nil, nil, err
	}
	if refusal != nil && refusal.ClientErrors.Count() > 0 {
		namedMethod := stringOf(order, models.SalesOrderFieldRequestedFulfillmentMethodId) != ""
		if !refusalStopsConfirm(namedMethod, &refusal.ClientErrors) {
			return nil, nil, nil
		}
		return nil, &refusal.ClientErrors, nil
	}
	if resolved == nil || resolved.Type != models.FulfillmentTypeKioskDispense {
		return nil, nil, nil
	}

	lines, err := outstandingLines(ctx, orderId)
	if err != nil {
		return nil, nil, err
	}
	if len(lines) == 0 {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation(models.SalesOrderFulfillmentSchemaName,
			ReasonFulfillmentNothingToDeliver,
			"this order owes no goods, so there is nothing for a kiosk to dispense"))
		return nil, vErrs, nil
	}

	// A confirm that was refused, or that died before stamping the order, leaves a fulfillment row
	// behind on an order that is still a draft. Reuse it rather than writing a second: the
	// fulfillment id IS the reservation's source reference in Inventory, so reusing the row reuses
	// the hold, and the retry becomes idempotent on both sides of the seam at once.
	fulfillmentId, itemIds, expiresAt, err := reusableFulfillment(ctx, orderId, lines)
	if err != nil {
		return nil, nil, err
	}
	if fulfillmentId == "" {
		fulfillmentId, itemIds, expiresAt, err = writeFulfillment(ctx, orderId, orgId, resolved, lines)
		if err != nil {
			return nil, nil, err
		}
	}

	// The hold is taken outside the write transaction above: Inventory commits in its own, and the
	// two cannot be made atomic across modules. The order of the two matters — the fulfillment row
	// exists first, so a reservation that succeeds always has something to attribute itself to, and
	// one that fails leaves a row the caller refuses on rather than an orphaned hold.
	held, vErrs, err := reserveFulfillment(
		ctx, fulfillmentId, orgId, resolved, lines, itemIds, expiresAt, reservations)
	if err != nil || vErrs != nil {
		return nil, vErrs, err
	}

	status := models.FulfillmentStatusReserved
	if !held {
		status = models.FulfillmentStatusPendingReservation
	}
	if err := setFulfillmentStatus(ctx, fulfillmentId, status); err != nil {
		return nil, nil, err
	}

	result := &CreateFulfillmentResult{
		FulfillmentId:  fulfillmentId,
		Status:         string(status),
		TargetOutletId: resolved.TargetOutletId,
	}
	if expiresAt != nil {
		result.ReservationExpiresAt = expiresAt.GoTime().Format(time.RFC3339)
	}
	return result, nil, nil
}

// writeFulfillment stores the header and its items in one transaction, so a half-written
// fulfillment cannot exist: a header with no items would reserve nothing and report itself ready.
func writeFulfillment(
	ctx corectx.Context,
	orderId, orgId string,
	resolved *ResolvedFulfillmentMethod,
	lines []itExt.FulfillmentLine,
) (fulfillmentId string, itemIds []string, expiresAt *model.ModelDateTime, err error) {
	id, err := model.NewId()
	if err != nil {
		return "", nil, nil, err
	}
	fulfillmentId = string(*id)
	expiresAt = reservationDeadline(resolved.ReservationTtlMinutes, time.Now().UTC())

	itemIds = make([]string, len(lines))
	for index := range lines {
		itemId, err := model.NewId()
		if err != nil {
			return "", nil, nil, err
		}
		itemIds[index] = string(*itemId)
	}

	err = withTransaction(ctx, models.SalesOrderFulfillmentSchemaName, func(tranxCtx corectx.Context) error {
		engineRepo, err := repoFor(models.SalesOrderFulfillmentSchemaName)
		if err != nil {
			return err
		}

		fields := dmodel.DynamicFields{
			models.SalesOrderFulfillmentFieldId:                  fulfillmentId,
			models.SalesOrderFulfillmentFieldSalesOrderId:        orderId,
			basemodel.FieldOrgId:                                 orgId,
			models.SalesOrderFulfillmentFieldFulfillmentMethodId: resolved.MethodId,

			// Every policy field below is a COPY. Nothing downstream reads the method again, so
			// archiving or editing it never changes what a live fulfillment promised.
			models.SalesOrderFulfillmentFieldFulfillmentType:         string(resolved.Type),
			models.SalesOrderFulfillmentFieldTargetOutletId:          resolved.TargetOutletId,
			models.SalesOrderFulfillmentFieldFulfillmentStatus:       string(models.FulfillmentStatusPendingReservation),
			models.SalesOrderFulfillmentFieldFailureAction:           string(resolved.FailureAction),
			models.SalesOrderFulfillmentFieldAllowTargetChange:       resolved.AllowTargetChange,
			models.SalesOrderFulfillmentFieldAllowPartialFulfillment: resolved.AllowPartialFulfillment,
		}
		if resolved.MaxAttempts != nil {
			fields[models.SalesOrderFulfillmentFieldMaxAttempts] = *resolved.MaxAttempts
		}
		if resolved.ReservationTtlMinutes != nil {
			fields[models.SalesOrderFulfillmentFieldReservationTtlMinutes] = *resolved.ReservationTtlMinutes
		}
		if expiresAt != nil {
			fields[models.SalesOrderFulfillmentFieldReservationExpiresAt] = *expiresAt
		}

		if _, err := engineRepo.Insert(tranxCtx, fields); err != nil {
			return errors.Wrap(err, "writing the order fulfillment")
		}
		return writeFulfillmentItems(tranxCtx, fulfillmentId, orgId, lines, itemIds)
	})
	if err != nil {
		return "", nil, nil, err
	}
	return fulfillmentId, itemIds, expiresAt, nil
}

func writeFulfillmentItems(
	ctx corectx.Context, fulfillmentId, orgId string, lines []itExt.FulfillmentLine, itemIds []string,
) error {
	engineRepo, err := repoFor(models.SalesOrderFulfillmentItemSchemaName)
	if err != nil {
		return err
	}

	for index, line := range lines {
		fields := dmodel.DynamicFields{
			models.SalesOrderFulfillmentItemFieldId:               itemIds[index],
			models.SalesOrderFulfillmentItemFieldFulfillmentId:    fulfillmentId,
			models.SalesOrderFulfillmentItemFieldSalesOrderLineId: line.SalesOrderLineId,
			models.SalesOrderFulfillmentItemFieldProductVariantId: line.ProductVariantId,
			models.SalesOrderFulfillmentItemFieldUomId:            line.UomId,
			models.SalesOrderFulfillmentItemFieldOrderedQty:       line.Quantity,
			models.SalesOrderFulfillmentItemFieldSourceLocationId: line.SourceLocationId,

			// Zeroed rather than omitted: the columns are NOT NULL, and nothing has been delivered
			// or refunded at the moment a fulfillment is created.
			models.SalesOrderFulfillmentItemFieldFulfilledQty: decimal.Zero,
			models.SalesOrderFulfillmentItemFieldRefundedQty:  decimal.Zero,
			models.SalesOrderFulfillmentItemFieldItemStatus:   string(models.FulfillmentItemStatusPending),

			basemodel.FieldOrgId: orgId,
		}
		if _, err := engineRepo.Insert(ctx, fields); err != nil {
			return errors.Wrap(err, "writing an order fulfillment item")
		}
	}
	return nil
}

// reserveFulfillment claims the stock and records what Inventory gave back against each item.
//
// A partial hold is refused rather than accepted. The customer is being promised one machine, and
// half the goods at that machine is not a smaller version of the promise — it is a different sale,
// which they should be allowed to decline before paying.
//
// Items are grouped by the location they come from, and each group is a hold of its own. A vending
// sale spans several slots, and Inventory holds stock at ONE location per reservation, so a single
// call could not express it. Grouping is also what makes a partial dispense attributable later: the
// slot that failed releases its own hold while the slots that succeeded keep theirs.
func reserveFulfillment(
	ctx corectx.Context,
	fulfillmentId, orgId string,
	resolved *ResolvedFulfillmentMethod,
	lines []itExt.FulfillmentLine,
	itemIds []string,
	expiresAt *model.ModelDateTime,
	reservations itExt.FulfillmentReservationExtService,
) (bool, *ft.ClientErrors, error) {
	if reservations == nil {
		// No port bound. The fulfillment stands unreserved rather than failing the confirm, matching
		// how the goods-issue path treats an absent inventory port.
		return false, nil, nil
	}

	groups := groupItemsByLocation(resolved.TargetLocationId, lines, itemIds)
	for _, group := range groups {
		response, err := reservations.ReserveForFulfillment(ctx, itExt.FulfillmentReservationRequest{
			FulfillmentId: reservationSourceId(fulfillmentId, group.LocationId, resolved.TargetLocationId),
			OrgId:         orgId,
			LocationId:    group.LocationId,
			ExpiresAt:     expiresAt,
			Items:         group.Items,
		})
		refusal, err := assertGroupReserved(ctx, fulfillmentId, response, err, reservations)
		if err != nil || refusal != nil {
			return false, refusal, err
		}
		if err := stampReservationRefs(ctx, group.ItemIds, response.InventoryReference); err != nil {
			return false, nil, err
		}
	}
	return true, nil, nil
}

// assertGroupReserved turns one group's answer into a decision, releasing EVERY hold this
// fulfillment has taken before refusing.
//
// Releasing all of them rather than only the group that failed is what the all-or-nothing promise
// requires: the earlier groups succeeded, and leaving their stock held for a sale that is being
// refused would make goods unsellable for a customer who never got them.
func assertGroupReserved(
	ctx corectx.Context,
	fulfillmentId string,
	response *itExt.FulfillmentReservationResponse,
	reserveErr error,
	reservations itExt.FulfillmentReservationExtService,
) (*ft.ClientErrors, error) {
	if reserveErr != nil {
		// A transport or database fault says nothing about whether the claim landed. Attempt the
		// release before propagating, so a reservation that did commit does not outlive the confirm
		// that was abandoned; the original error is what the caller hears either way.
		releaseAfterFailedReserve(ctx, fulfillmentId, reservations)
		return nil, reserveErr
	}
	if response == nil || !response.Accepted {
		// Nothing was necessarily claimed, but a refusal can still land after a partial claim, so the
		// release runs on this branch too. Releasing a hold that never existed is a no-op.
		releaseAfterFailedReserve(ctx, fulfillmentId, reservations)
		return reservationRefusal(response, ReasonFulfillmentReserveFailed), nil
	}
	if !response.FullyReserved {
		// Inventory holds the part it could claim. Giving it back before refusing is what stops a
		// declined sale sitting on stock: the TTL sweep would eventually reclaim it, but a method
		// with no reservation_ttl_minutes is never swept, so the goods would be held forever.
		releaseAfterFailedReserve(ctx, fulfillmentId, reservations)
		return reservationRefusal(response, ReasonFulfillmentInsufficientStock), nil
	}
	return nil, nil
}

// reservationGroup is the items of one fulfillment that come from one location.
type reservationGroup struct {
	LocationId string
	ItemIds    []string
	Items      []itExt.FulfillmentReservationItem
}

// groupItemsByLocation splits a fulfillment's items by where their stock sits, preserving the order
// the lines were given so a group's items stay in a stable, reproducible order.
//
// A line naming no location falls to the fulfillment's single target, which is how every target
// without addressable slots behaves and why this change leaves those flows untouched.
func groupItemsByLocation(
	targetLocationId string, lines []itExt.FulfillmentLine, itemIds []string,
) []reservationGroup {
	groups := make([]reservationGroup, 0, 1)
	indexOf := make(map[string]int, 1)

	for index, line := range lines {
		locationId := line.SourceLocationId
		if locationId == "" {
			locationId = targetLocationId
		}
		at, seen := indexOf[locationId]
		if !seen {
			at = len(groups)
			indexOf[locationId] = at
			groups = append(groups, reservationGroup{LocationId: locationId})
		}
		groups[at].ItemIds = append(groups[at].ItemIds, itemIds[index])
		groups[at].Items = append(groups[at].Items, itExt.FulfillmentReservationItem{
			FulfillmentItemId: itemIds[index],
			ProductVariantId:  line.ProductVariantId,
			UomId:             line.UomId,
			Quantity:          line.Quantity,
			SourceLocationId:  locationId,
		})
	}
	return groups
}

// stampReservationRefs records the hold against every item, so an operator tracing one item can
// find the inventory document without searching by origin.
func stampReservationRefs(ctx corectx.Context, itemIds []string, reference string) error {
	if reference == "" {
		return nil
	}
	for _, itemId := range itemIds {
		record, err := loadRecord(ctx, models.SalesOrderFulfillmentItemSchemaName,
			models.SalesOrderFulfillmentItemFieldId, itemId)
		if err != nil {
			return err
		}
		if record == nil {
			continue
		}
		err = writeChanges(ctx, models.SalesOrderFulfillmentItemSchemaName, record, dmodel.DynamicFields{
			models.SalesOrderFulfillmentItemFieldInventoryReservationRef: reference,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// setFulfillmentStatus moves the header. It goes through writeChanges rather than the resource
// service because fulfillment_status is declared no_update: these services are the only sanctioned
// way to move it, and a client that could set it would be able to declare goods delivered.
func setFulfillmentStatus(
	ctx corectx.Context, fulfillmentId string, status models.FulfillmentStatus,
) error {
	record, err := loadRecord(ctx, models.SalesOrderFulfillmentSchemaName,
		models.SalesOrderFulfillmentFieldId, fulfillmentId)
	if err != nil {
		return err
	}
	if record == nil {
		return errors.New("the fulfillment to update no longer exists: " + fulfillmentId)
	}
	return writeChanges(ctx, models.SalesOrderFulfillmentSchemaName, record, dmodel.DynamicFields{
		models.SalesOrderFulfillmentFieldFulfillmentStatus: string(status),
	})
}

// reservationDeadline turns a TTL in minutes into an absolute instant. Null TTL means the hold does
// not lapse on a timer, which suits goods dispensed seconds after payment; storing a deadline for
// those would expire a sale that was already complete.
func reservationDeadline(ttlMinutes *int32, now time.Time) *model.ModelDateTime {
	if ttlMinutes == nil || *ttlMinutes <= 0 {
		return nil
	}
	deadline := model.ModelDateTime(now.Add(time.Duration(*ttlMinutes) * time.Minute))
	return &deadline
}

// reservationRefusal carries Inventory's own words to the customer-facing refusal, so an operator
// can tell whether to wait for a restock, offer another kiosk, or do nothing.
func reservationRefusal(
	response *itExt.FulfillmentReservationResponse, reason string,
) *ft.ClientErrors {
	message := "the target could not hold the goods for this order"
	if response != nil && response.FailureReason != "" {
		message = response.FailureReason
	}
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(models.SalesOrderFulfillmentSchemaName, reason, message))
	return vErrs
}

func hasReason(vErrs *ft.ClientErrors, reason string) bool {
	if vErrs == nil {
		return false
	}
	for _, item := range *vErrs {
		if item.Key == reason {
			return true
		}
	}
	return false
}

// refusalStopsConfirm decides whether a failed resolution aborts the confirm or means "this is not
// a kiosk sale", which is THE guard keeping every other channel on its original path.
//
// Exactly one refusal is survivable, and only for an order that named no method: "no method could be
// resolved" then says the deployment simply has no kiosk policy for this sale, which is the ordinary
// case for every non-kiosk order ever confirmed. Everything else aborts — an order that asked for a
// method by id must hear why it cannot have it, and a target that is unusable is a real problem
// whether or not the method was named.
//
// Note what this means operationally: once a channel or sales point carries a
// default_fulfillment_method_id, orders that named nothing DO resolve, and a target refusal then
// stops a confirm that used to succeed. That widening is configuration-driven, not code-driven.
func refusalStopsConfirm(namedMethod bool, vErrs *ft.ClientErrors) bool {
	if namedMethod {
		return true
	}
	return !hasReason(vErrs, ReasonMethodUnresolved)
}

// releaseAfterFailedReserve gives back whatever a failed reservation managed to claim.
//
// Its own failure is deliberately swallowed. This runs on a path that is already refusing the
// confirm, and replacing a refusal the customer can act on ("this kiosk cannot supply that") with a
// 500 about Inventory would tell them nothing useful and lose the real reason. What is left behind
// in that case is a hold with a TTL, which the sweep reclaims — the case this function exists to
// prevent is the one with no TTL at all, where nothing else ever would.
func releaseAfterFailedReserve(
	ctx corectx.Context, fulfillmentId string, reservations itExt.FulfillmentReservationExtService,
) {
	_ = releaseEveryHold(ctx, fulfillmentId, reservations)
}

// reusableFulfillment finds a fulfillment of this order that a retried confirm may take over, and
// returns its id, its item ids in the order the lines were given, and its deadline.
//
// Only a `pending_reservation` fulfillment qualifies. That is the state a refused or interrupted
// confirm leaves behind, and it holds nothing: anything further along has either been reserved for a
// confirmed order or deliberately cancelled, and quietly adopting one of those would let a second
// confirm take over a live sale.
//
// The items must correspond exactly to the lines being fulfilled. If the draft was edited between
// the two attempts the old row describes a different sale, so it is not reused — writing a fresh one
// is correct, and the stale row stays visible rather than being silently repurposed.
func reusableFulfillment(
	ctx corectx.Context, orderId string, lines []itExt.FulfillmentLine,
) (fulfillmentId string, itemIds []string, expiresAt *model.ModelDateTime, err error) {
	existing, err := FulfillmentsOfOrder(ctx, orderId)
	if err != nil {
		return "", nil, nil, err
	}

	for _, record := range existing {
		status := models.FulfillmentStatus(
			stringOf(record, models.SalesOrderFulfillmentFieldFulfillmentStatus))
		if status != models.FulfillmentStatusPendingReservation {
			continue
		}

		candidateId := stringOf(record, models.SalesOrderFulfillmentFieldId)
		items, err := ItemsOfFulfillment(ctx, candidateId)
		if err != nil {
			return "", nil, nil, err
		}

		matched := matchItemsToLines(items, lines)
		if matched == nil {
			continue
		}
		return candidateId, matched,
			dateTimeOf(record, models.SalesOrderFulfillmentFieldReservationExpiresAt), nil
	}
	return "", nil, nil, nil
}

// matchItemsToLines pairs stored items with the lines being fulfilled, returning the item ids in the
// lines' own order, or nil when they do not correspond.
//
// Matching is by order line AND variant AND quantity: the line id alone would accept a row whose
// quantity was edited between attempts, and reusing that would reserve the old amount for the new
// sale. One line may appear twice, so each stored item is consumed at most once.
func matchItemsToLines(items []dmodel.DynamicFields, lines []itExt.FulfillmentLine) []string {
	if len(items) != len(lines) {
		return nil
	}

	used := make([]bool, len(items))
	matched := make([]string, len(lines))
	for lineIndex, line := range lines {
		found := false
		for itemIndex, item := range items {
			if used[itemIndex] {
				continue
			}
			sameLine := stringOf(item, models.SalesOrderFulfillmentItemFieldSalesOrderLineId) == line.SalesOrderLineId
			sameVariant := stringOf(item, models.SalesOrderFulfillmentItemFieldProductVariantId) == line.ProductVariantId
			sameQty := decimalOf(item, models.SalesOrderFulfillmentItemFieldOrderedQty).Equal(line.Quantity)
			if !sameLine || !sameVariant || !sameQty {
				continue
			}
			used[itemIndex] = true
			matched[lineIndex] = stringOf(item, models.SalesOrderFulfillmentItemFieldId)
			found = true
			break
		}
		if !found {
			return nil
		}
	}
	return matched
}
