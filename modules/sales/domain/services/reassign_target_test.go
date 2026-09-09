package services

import (
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The preconditions of moving a delivery, and the rule that decides what actually moves.

func reassignableFulfillment(status models.FulfillmentStatus, allowChange bool) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.SalesOrderFulfillmentFieldId:                "FU1",
		models.SalesOrderFulfillmentFieldFulfillmentStatus: string(status),
		models.SalesOrderFulfillmentFieldFulfillmentType:   string(models.FulfillmentTypeKioskDispense),
		models.SalesOrderFulfillmentFieldTargetOutletId:    "SP1",
		models.SalesOrderFulfillmentFieldAllowTargetChange: allowChange,
	}
}

// THE policy guard. A sale made at the machine the customer was standing at promised no second
// machine, and an operator editing the catalogue afterwards must not be able to grant one — which is
// why this reads the fulfillment's own snapshot rather than the live method.
func TestAMethodThatForbidsTargetChangeRefuses(t *testing.T) {
	fulfillment := reassignableFulfillment(models.FulfillmentStatusReserved, false)
	params := ReassignTargetParams{FulfillmentId: "FU1", ToOutletId: "SP2"}

	vErrs := assertReassignable(fulfillment, params, "SP1")
	if vErrs == nil || !hasReason(vErrs, ReasonTargetChangeNotAllowed) {
		t.Errorf("a method forbidding target change must refuse, got %v", vErrs)
	}
}

func TestAMethodThatAllowsTargetChangePasses(t *testing.T) {
	fulfillment := reassignableFulfillment(models.FulfillmentStatusReserved, true)
	params := ReassignTargetParams{FulfillmentId: "FU1", ToOutletId: "SP2"}

	if vErrs := assertReassignable(fulfillment, params, "SP1"); vErrs != nil {
		t.Errorf("a permitted move must pass the policy gate, got %v", *vErrs)
	}
}

// A finished or called-off delivery has nowhere to move to.
func TestATerminalFulfillmentCannotBeMoved(t *testing.T) {
	for _, status := range []models.FulfillmentStatus{
		models.FulfillmentStatusCompleted,
		models.FulfillmentStatusCancelled,
	} {
		t.Run(string(status), func(t *testing.T) {
			fulfillment := reassignableFulfillment(status, true)
			params := ReassignTargetParams{FulfillmentId: "FU1", ToOutletId: "SP2"}

			vErrs := assertReassignable(fulfillment, params, "SP1")
			if vErrs == nil || !hasReason(vErrs, ReasonTargetChangeTerminal) {
				t.Errorf("a %s fulfillment must not be movable, got %v", status, vErrs)
			}
		})
	}
}

// Moving to where it already is would take a real reservation apart and rebuild it for no reason.
func TestMovingToTheSameOutletIsRefused(t *testing.T) {
	fulfillment := reassignableFulfillment(models.FulfillmentStatusReserved, true)
	params := ReassignTargetParams{FulfillmentId: "FU1", ToOutletId: "SP1"}

	vErrs := assertReassignable(fulfillment, params, "SP1")
	if vErrs == nil || !hasReason(vErrs, ReasonTargetChangeSameOutlet) {
		t.Errorf("moving to the current target must be refused, got %v", vErrs)
	}
}

func TestAMoveWithNoDestinationIsRefused(t *testing.T) {
	fulfillment := reassignableFulfillment(models.FulfillmentStatusReserved, true)
	params := ReassignTargetParams{FulfillmentId: "FU1"}

	vErrs := assertReassignable(fulfillment, params, "SP1")
	if vErrs == nil || !hasReason(vErrs, ReasonTargetRequired) {
		t.Errorf("a move with no destination must be refused, got %v", vErrs)
	}
}

// A fulfillment waiting on the customer is exactly the one they are most likely to want moved, so
// waiting_customer_action must remain movable.
func TestAFulfillmentWaitingOnTheCustomerCanBeMoved(t *testing.T) {
	fulfillment := reassignableFulfillment(models.FulfillmentStatusWaitingCustomerAction, true)
	params := ReassignTargetParams{FulfillmentId: "FU1", ToOutletId: "SP2"}

	if vErrs := assertReassignable(fulfillment, params, "SP1"); vErrs != nil {
		t.Errorf("a customer choosing another kiosk is the main use of this, got %v", *vErrs)
	}
}

// THE rule of what moves. Only what is still owed is reallocated: goods already dispensed are the
// customer's and are not at the old kiosk to move, so asking Inventory to shift them would claim
// stock for a delivery that already happened.
func TestOnlyOutstandingQuantityIsMoved(t *testing.T) {
	items := []dmodel.DynamicFields{
		{
			models.SalesOrderFulfillmentItemFieldId:               "FI1",
			models.SalesOrderFulfillmentItemFieldProductVariantId: "V1",
			models.SalesOrderFulfillmentItemFieldUomId:            "U1",
			models.SalesOrderFulfillmentItemFieldOrderedQty:       qtyOf("5"),
			models.SalesOrderFulfillmentItemFieldFulfilledQty:     qtyOf("2"),
			models.SalesOrderFulfillmentItemFieldRefundedQty:      qtyOf("0"),
		},
	}

	total, reservationItems := outstandingReservationItems(items)
	if !total.Equal(qtyOf("3")) {
		t.Errorf("5 ordered with 2 delivered leaves 3 to move, got %s", total)
	}
	if len(reservationItems) != 1 {
		t.Fatalf("expected one item to move, got %d", len(reservationItems))
	}
	if !reservationItems[0].Quantity.Equal(qtyOf("3")) {
		t.Errorf("the moved quantity must be what is still owed, got %s", reservationItems[0].Quantity)
	}
}

// A fully settled item is not carried to the new location at all.
func TestASettledItemIsNotMoved(t *testing.T) {
	items := []dmodel.DynamicFields{
		{
			models.SalesOrderFulfillmentItemFieldId:           "FI1",
			models.SalesOrderFulfillmentItemFieldOrderedQty:   qtyOf("2"),
			models.SalesOrderFulfillmentItemFieldFulfilledQty: qtyOf("2"),
		},
		{
			models.SalesOrderFulfillmentItemFieldId:           "FI2",
			models.SalesOrderFulfillmentItemFieldOrderedQty:   qtyOf("1"),
			models.SalesOrderFulfillmentItemFieldFulfilledQty: qtyOf("0"),
		},
	}

	total, reservationItems := outstandingReservationItems(items)
	if !total.Equal(qtyOf("1")) {
		t.Errorf("only the undelivered unit moves, got %s", total)
	}
	if len(reservationItems) != 1 || reservationItems[0].FulfillmentItemId != "FI2" {
		t.Errorf("the delivered item must not be moved, got %v", reservationItems)
	}
}

// A refunded quantity is settled too: the customer was paid back, so there is nothing to promise
// them from anywhere.
func TestARefundedQuantityIsNotMoved(t *testing.T) {
	items := []dmodel.DynamicFields{
		{
			models.SalesOrderFulfillmentItemFieldId:           "FI1",
			models.SalesOrderFulfillmentItemFieldOrderedQty:   qtyOf("3"),
			models.SalesOrderFulfillmentItemFieldFulfilledQty: qtyOf("1"),
			models.SalesOrderFulfillmentItemFieldRefundedQty:  qtyOf("2"),
		},
	}

	total, _ := outstandingReservationItems(items)
	if !total.IsZero() {
		t.Errorf("1 delivered and 2 refunded of 3 owes nothing, got %s", total)
	}
}

// Who made the change is read from the request's own authentication, never from a payload: a caller
// able to name its own actor could attribute its changes to somebody else.
func TestTheActorIsReadFromTheRequest(t *testing.T) {
	ctx := corectx.NewRequestContext(t.Context())
	ctx.SetPermissions(corectx.ContextPermissions{
		Principal: corectx.Principal{Kind: corectx.PrincipalKindUser, Id: model.Id("U1")},
	})

	actorType, actorId := describeActor(ctx)
	if actorType != models.TargetChangeActorTypeUser || actorId != "U1" {
		t.Errorf("a person's change must be attributed to them, got %s/%s", actorType, actorId)
	}
}

// A kiosk is a workload, and "the kiosk moved it" is a different answer to a customer than "a
// support agent moved it".
func TestAServiceActorIsRecordedAsAService(t *testing.T) {
	ctx := corectx.NewRequestContext(t.Context())
	ctx.SetPermissions(corectx.ContextPermissions{
		Principal: corectx.Principal{Kind: corectx.PrincipalKindService, Id: model.Id("KIOSK1")},
	})

	actorType, _ := describeActor(ctx)
	if actorType != models.TargetChangeActorTypeService {
		t.Errorf("a workload must be recorded as a service, got %s", actorType)
	}
}

// An unauthenticated execution is a sweep or a job. Inventing a user for it would put a person's
// name against a change no person made.
func TestAnUnauthenticatedChangeIsRecordedAsSystem(t *testing.T) {
	actorType, actorId := describeActor(nil)
	if actorType != models.TargetChangeActorTypeSystem || actorId != "" {
		t.Errorf("a change with no principal is the system's, got %s/%s", actorType, actorId)
	}

	ctx := corectx.NewRequestContext(t.Context())
	if actorType, _ := describeActor(ctx); actorType != models.TargetChangeActorTypeSystem {
		t.Errorf("an empty principal is the system, got %s", actorType)
	}
}
