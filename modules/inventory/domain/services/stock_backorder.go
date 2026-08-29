package services

import (
	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// Backorder: what happens to the part of a demand a validate did not deliver.
//
// The original transfer's demand is NEVER rewritten to what was processed. A transfer that asked
// for 100 and shipped 70 stays that way, and the outstanding 30 becomes a new document; rewriting
// the demand would erase the fact that 30 were promised and not delivered.

// BackorderDecision is what to do with an unprocessed remainder.
type BackorderDecision string

const (
	// BackorderNone means nothing was left over, so the question does not arise.
	BackorderNone BackorderDecision = "none"

	// BackorderCreate raises a new transfer for the remainder.
	BackorderCreate BackorderDecision = "create"

	// BackorderDrop abandons the remainder: the demand will not be met and no document survives it.
	BackorderDrop BackorderDecision = "drop"
)

// DecideBackorder applies the transfer's snapshot policy to what the moves actually processed. The
// `ask` policy requires an answer: defaulting it either way would make the setting meaningless, so
// a missing decision is a client error.
func DecideBackorder(
	policy string, outcomes []moveOutcome, createBackorder *bool,
) (BackorderDecision, *ft.ClientErrors) {
	vErrs := ft.NewClientErrors()

	if !hasShortfall(outcomes) {
		return BackorderNone, vErrs
	}

	switch policy {
	case models.StockBackorderPolicyAlways:
		return BackorderCreate, vErrs
	case models.StockBackorderPolicyNever:
		return BackorderDrop, vErrs
	case models.StockBackorderPolicyAsk:
		if createBackorder == nil {
			vErrs.Append(*ft.NewBusinessViolation(
				models.StockTransferSchemaName,
				"stock_transfer.backorder_decision_required",
				"this transfer's backorder policy is 'ask': send create_backorder to say whether the "+
					"undelivered quantity should become a new transfer",
			))
			return BackorderNone, vErrs
		}
		if *createBackorder {
			return BackorderCreate, vErrs
		}
		return BackorderDrop, vErrs
	default:
		// An unknown policy is a data problem, but refusing beats guessing: dropping a remainder that
		// should have been backordered loses a commitment.
		vErrs.Append(*ft.NewBusinessViolation(
			models.StockTransferSchemaName,
			"stock_transfer.unknown_backorder_policy",
			"unknown backorder policy '"+policy+"'",
		))
		return BackorderNone, vErrs
	}
}

func hasShortfall(outcomes []moveOutcome) bool {
	for _, outcome := range outcomes {
		if outcome.Shortfall().GreaterThan(decimal.Zero) {
			return true
		}
	}
	return false
}

// createBackorderTransfer raises a new draft transfer carrying the undelivered quantities. A new
// document, not a reopened one: the original stays done, and backorder_of_id makes the chain
// traceable in both directions.
func createBackorderTransfer(
	ctx corectx.Context, operation *transferOperationContext, outcomes []moveOutcome,
) error {
	original := operation.Transfer
	operationCode := derefString(original.GetOperationCode())

	transferNumber, err := generateTransferNumber(operationCode)
	if err != nil {
		return err
	}

	created, err := operation.TransferEngine.ResourceRepository().Insert(ctx, dmodel.DynamicFields{
		models.StockTransferFieldTransferNumber:        transferNumber,
		models.StockTransferFieldOperationTypeId:       derefString(original.GetOperationTypeId()),
		models.StockTransferFieldOperationCode:         operationCode,
		models.StockTransferFieldSourceLocationId:      derefString(original.GetSourceLocationId()),
		models.StockTransferFieldDestinationLocationId: derefString(original.GetDestinationLocationId()),
		models.StockTransferFieldStatus:                models.StockTransferStatusDraft,
		// Policies are copied from the original rather than re-read from the operation type, so the
		// backorder behaves as its parent did even if the type has since changed.
		models.StockTransferFieldReservationMethod: derefString(original.GetReservationMethod()),
		models.StockTransferFieldBackorderPolicy:   derefString(original.GetBackorderPolicy()),
		models.StockTransferFieldShippingPolicy:    derefString(original.GetShippingPolicy()),
		models.StockTransferFieldBackorderOfId:     derefString(original.GetId()),
		models.StockTransferFieldOrgId:             derefString(original.GetOrgId()),
	})
	if err != nil {
		return errors.Wrap(err, "createBackorderTransfer")
	}
	_ = created

	backorderId, err := findTransferByNumber(ctx, operation, derefString(original.GetOrgId()), transferNumber)
	if err != nil {
		return err
	}
	return copyShortfallMoves(ctx, operation, outcomes, backorderId)
}

// findTransferByNumber reads back the transfer just inserted, by the number that is unique per org.
func findTransferByNumber(
	ctx corectx.Context, operation *transferOperationContext, orgId, transferNumber string,
) (string, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.StockTransferFieldOrgId, dmodel.Equals, orgId),
		*dmodel.NewSearchNode().NewCondition(
			models.StockTransferFieldTransferNumber, dmodel.Equals, transferNumber),
	)
	found, err := operation.TransferEngine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  1,
	})
	if err != nil {
		return "", errors.Wrap(err, "findTransferByNumber")
	}
	if !found.HasData || len(found.Data.Items) == 0 {
		return "", errors.New("the backorder transfer could not be read back after being created")
	}
	return derefString(models.NewStockTransferFrom(found.Data.Items[0]).GetId()), nil
}

// copyShortfallMoves gives the backorder a move for each quantity the original did not deliver.
func copyShortfallMoves(
	ctx corectx.Context, operation *transferOperationContext, outcomes []moveOutcome, backorderId string,
) error {
	byId := indexMovesById(operation.Moves)

	for _, outcome := range outcomes {
		shortfall := outcome.Shortfall()
		if shortfall.LessThanOrEqual(decimal.Zero) {
			continue
		}
		source, ok := byId[outcome.MoveId]
		if !ok {
			continue
		}

		_, err := operation.MoveEngine.ResourceRepository().Insert(ctx, dmodel.DynamicFields{
			models.StockMoveFieldTransferId:            backorderId,
			models.StockMoveFieldProductVariantId:      derefString(source.GetProductVariantId()),
			models.StockMoveFieldDemandQuantity:        shortfall.String(),
			models.StockMoveFieldBaseDemandQuantity:    shortfall.String(),
			models.StockMoveFieldSourceLocationId:      derefString(source.GetSourceLocationId()),
			models.StockMoveFieldDestinationLocationId: derefString(source.GetDestinationLocationId()),
			models.StockMoveFieldStatus:                models.StockMoveStatusDraft,
			// The backorder's move points at the move it carries the remainder of, so a split demand
			// can be followed back to the one originally raised.
			models.StockMoveFieldOriginMoveId: outcome.MoveId,
			models.StockMoveFieldOrgId:        derefString(source.GetOrgId()),
		})
		if err != nil {
			return errors.Wrap(err, "copyShortfallMoves")
		}
	}
	return nil
}

func indexMovesById(moves []dmodel.DynamicFields) map[string]*models.StockMove {
	byId := make(map[string]*models.StockMove, len(moves))
	for _, item := range moves {
		move := models.NewStockMoveFrom(item)
		byId[derefString(move.GetId())] = move
	}
	return byId
}
