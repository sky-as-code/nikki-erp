package services

import (
	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// What Inventory does when an executor reports what physically happened.
//
// The caller — a vending module relaying a machine's dispense result — says which quantities left
// and which did not. Inventory decides the consequence, and that division matters: a machine
// reporting "the motor turned but nothing reached the tray" is reporting a fact, not a conclusion.
// Whether those goods are still sellable, jammed, or need a human to look is Inventory's judgement,
// and a caller that decided it would be writing stock levels from outside.
//
// The shape of the operation follows from what a partial result means. A hold was taken for the
// whole demand; part of it left the building and part did not. So the demand is trimmed to what
// actually moved, that much is validated — which is the one path that decrements on-hand — and
// whatever was held for the rest is released, because nothing is going to consume it.

// ApplyFulfillmentResult records the physical outcome of an executor's attempt against a demand's
// reservation.
//
// Idempotent by event id: the same event applied twice consumes stock once. That is not a nicety —
// the caller relays events from a message broker that may deliver one twice, and a second
// consumption would take goods off the shelf that no customer received.
func (this *StockTransferDomainServiceImpl) ApplyFulfillmentResult(
	ctx corectx.Context, request itStock.FulfillmentResultRequest,
) (*itStock.FulfillmentResultResponse, error) {
	if result := assertFulfillmentResultWellFormed(request); result != nil {
		return result, nil
	}

	var outcome *itStock.FulfillmentResultResponse

	err := withTransferTransaction(ctx, func(tranxCtx corectx.Context) error {
		transferIds, err := this.findAllReservationsBySource(
			tranxCtx, request.SourceType, request.SourceId)
		if err != nil {
			return err
		}
		if len(transferIds) == 0 {
			// Either the event is a replay whose hold is already closed, or it names a demand that
			// never held anything. Both are answered the same way and neither is a fault: the
			// caller's intent — "this much left, this much did not" — needs nothing further done.
			outcome = &itStock.FulfillmentResultResponse{
				InventoryResultRef: request.EventId,
				AlreadyApplied:     true,
			}
			return nil
		}

		dispensed := dispensedByItem(request)

		for _, transferId := range transferIds {
			operation, err := loadTransferOperation(tranxCtx, transferId)
			if err != nil {
				return err
			}
			if operation == nil {
				continue
			}

			if err := trimMovesToDispensed(tranxCtx, operation, dispensed); err != nil {
				return err
			}

			// Validate is what actually moves the goods, and it consumes exactly the reservations
			// still standing on each move — which, after the trim, is what came out of the machine.
			// The event id is its idempotency key, so a replayed event finds the transfer already
			// validated and is answered from the stored result rather than moving stock twice.
			if _, err := this.Validate(tranxCtx, transferId, request.EventId, ptrOf(false)); err != nil {
				return err
			}
		}

		outcome = &itStock.FulfillmentResultResponse{InventoryResultRef: request.EventId}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return outcome, nil
}

// dispensedByItem indexes what actually left, by the caller's own line id.
//
// Keyed on the source item rather than on the product: one demand can name the same variant on two
// lines, and matching by product would attribute one line's success to the other.
func dispensedByItem(request itStock.FulfillmentResultRequest) map[string]decimal.Decimal {
	dispensed := make(map[string]decimal.Decimal, len(request.SuccessfulItems))
	for _, item := range request.SuccessfulItems {
		dispensed[item.SourceItemId] = dispensed[item.SourceItemId].Add(item.Quantity)
	}
	return dispensed
}

// trimMovesToDispensed cuts every move's demand down to the quantity that physically moved, and
// gives back the reservation held for the rest.
//
// Trimming before validating is what makes a partial result honest. Validate consumes what a move
// still has reserved, so a move left at its full demand would consume stock for goods that never
// reached the customer. A move that dispensed nothing is trimmed to zero and released entirely: its
// goods are still on the shelf, whatever the machine did with them, and Inventory's own disposition
// rules — not this function — decide whether they are sellable.
func trimMovesToDispensed(
	ctx corectx.Context, operation *transferOperationContext, dispensed map[string]decimal.Decimal,
) error {
	for _, item := range operation.Moves {
		move := models.NewStockMoveFrom(item)
		if !IsMoveOpen(derefString(move.GetStatus())) {
			continue
		}

		sourceItemId := derefString(move.GetSourceItemId())
		delivered := dispensed[sourceItemId]
		demand := orZero(move.GetBaseDemandQuantity())

		if delivered.GreaterThanOrEqual(demand) {
			// Everything asked for came out; the move stands as it is.
			continue
		}

		// Release the whole hold, then re-reserve only what left. Reducing the reserved lines in
		// place would mean editing a line to a quantity it never claimed; releasing and re-claiming
		// keeps every line a true record of a claim that was made.
		if err := releaseMoveLines(ctx, operation, *move); err != nil {
			return err
		}

		if err := writeMoveDemand(ctx, operation, *move, delivered); err != nil {
			return err
		}
		if delivered.LessThanOrEqual(decimal.Zero) {
			// Nothing to consume. The move is closed by the validate that follows, having moved
			// nothing, which is exactly what happened.
			continue
		}
		if _, err := reserveOneMove(ctx, operation, *move); err != nil {
			return err
		}
	}
	return nil
}

// writeMoveDemand rewrites what a move is asking for, so validate consumes that much and no more.
//
// Both quantity columns move together: reservation and validation read the BASE quantity, while the
// display quantity is what a human sees, and leaving them disagreeing would show a document that
// claims to have moved more than it did.
func writeMoveDemand(
	ctx corectx.Context, operation *transferOperationContext, move models.StockMove, quantity decimal.Decimal,
) error {
	update := dmodel.DynamicFields{
		models.StockMoveFieldId:                 derefString(move.GetId()),
		models.StockMoveFieldDemandQuantity:     quantity.String(),
		models.StockMoveFieldBaseDemandQuantity: quantity.String(),
	}
	_, err := operation.MoveEngine.ResourceRepository().Update(ctx, update)
	return errors.Wrap(err, "writeMoveDemand")
}

// assertFulfillmentResultWellFormed refuses a result that cannot be attributed, before any stock is
// touched. Without an event id the operation could not be made idempotent, and a replayed delivery
// would consume the goods a second time.
func assertFulfillmentResultWellFormed(
	request itStock.FulfillmentResultRequest,
) *itStock.FulfillmentResultResponse {
	vErrs := ft.NewClientErrors()
	refuse := func(key, message string) {
		vErrs.Append(*ft.NewBusinessViolation(models.StockTransferSchemaName, key, message))
	}

	if request.EventId == "" {
		refuse("stock_transfer.event_id_required",
			"a fulfillment result must carry the id of the event reporting it, which is what makes "+
				"a redelivered event consume stock only once")
	}
	if request.SourceType == "" || request.SourceId == "" {
		refuse("stock_transfer.source_required",
			"a fulfillment result must name the demand it reports on")
	}
	for index, item := range request.SuccessfulItems {
		if item.Quantity.LessThan(decimal.Zero) {
			refuse("stock_transfer.quantity_must_not_be_negative",
				"successful item "+decimal.NewFromInt(int64(index+1)).String()+
					" reports a negative quantity")
		}
	}
	for index, item := range request.FailedItems {
		if item.Quantity.LessThan(decimal.Zero) {
			refuse("stock_transfer.quantity_must_not_be_negative",
				"failed item "+decimal.NewFromInt(int64(index+1)).String()+
					" reports a negative quantity")
		}
	}

	if vErrs.Count() == 0 {
		return nil
	}
	return &itStock.FulfillmentResultResponse{ClientErrors: *vErrs}
}
