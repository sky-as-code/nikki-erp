package services

import (
	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"

	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
)

// The transfer's create and update carry its moves as a delta, following the nested-delta pattern
// of the IAM role: the transfer row is written first, so its etag guards the delta, then the moves
// through the stock-move domain service, so each move's own validation and audit apply. Any refusal
// rolls the transfer change back too.

const (
	movesFieldAdd    = itStock.TransferFieldMoves + ".add"
	movesFieldUpdate = itStock.TransferFieldMoves + ".update"
	movesFieldRemove = itStock.TransferFieldMoves + ".remove"
)

// liftMovesOrRefuse lifts the moves delta off the command and checks its shape. A nil delta means
// the command carries no move change.
func liftMovesOrRefuse(params dmodel.DynamicFields) (*itStock.TransferMovesDelta, ft.ClientErrors) {
	delta, malformed := itStock.LiftMovesDelta(params)
	if malformed {
		return nil, ft.ClientErrors{*ft.NewValidationError(
			itStock.TransferFieldMoves, ft.ErrorKey("err_malformed_request"),
			"moves must be {add: [{product_variant_id, uom_id, demand_quantity}], "+
				"update: [{id, etag, product_variant_id, uom_id, demand_quantity}], remove: [ids]}",
		)}
	}
	if delta == nil || delta.IsEmpty() {
		return nil, nil
	}
	return delta, validateMovesDeltaShape(*delta)
}

func validateMovesDeltaShape(delta itStock.TransferMovesDelta) ft.ClientErrors {
	cErrs := validateMovesDeltaSize(delta)
	if cErrs.Count() > 0 {
		return cErrs
	}
	for i, item := range delta.Add {
		cErrs = append(cErrs, validateMoveAdd(item, valuesDeltaRowPath(movesFieldAdd, i))...)
	}
	listed := make(map[string]bool, len(delta.Update)+len(delta.Remove))
	for i, item := range delta.Update {
		field := valuesDeltaRowPath(movesFieldUpdate, i)
		cErrs = append(cErrs, validateMoveUpdate(item, field)...)
		cErrs = append(cErrs, refuseMoveListedTwice(listed, string(item.Id), field)...)
	}
	for i, moveId := range delta.Remove {
		field := valuesDeltaRowPath(movesFieldRemove, i)
		if moveId == "" {
			cErrs = append(cErrs, *dmodel.NewMissingFieldErr(field))
			continue
		}
		cErrs = append(cErrs, refuseMoveListedTwice(listed, string(moveId), field)...)
	}
	return cErrs
}

func validateMovesDeltaSize(delta itStock.TransferMovesDelta) ft.ClientErrors {
	var out ft.ClientErrors
	maxItems := itStock.TransferMovesDeltaMaxItems
	vars := map[string]any{"max": maxItems}
	lists := map[string]int{
		movesFieldAdd: len(delta.Add), movesFieldUpdate: len(delta.Update), movesFieldRemove: len(delta.Remove),
	}
	for _, field := range []string{movesFieldAdd, movesFieldUpdate, movesFieldRemove} {
		if lists[field] > maxItems {
			out = append(out, *ft.NewValidationError(field, ft.ErrorKey("stock_transfer.too_many_moves", "inventory"),
				"at most {{max}} moves can be sent in one list", vars))
		}
	}
	return out
}

func validateMoveAdd(item itStock.TransferMoveAdd, field string) ft.ClientErrors {
	var out ft.ClientErrors
	if item.ProductVariantId == "" {
		out = append(out, *dmodel.NewMissingFieldErr(field + "." + models.StockMoveFieldProductVariantId))
	}
	if item.DemandQuantity == nil {
		out = append(out, *dmodel.NewMissingFieldErr(field + "." + models.StockMoveFieldDemandQuantity))
	} else {
		out = append(out, validateMoveQuantity(*item.DemandQuantity, field)...)
	}
	return out
}

func validateMoveUpdate(item itStock.TransferMoveUpdate, field string) ft.ClientErrors {
	var out ft.ClientErrors
	if item.Id == "" {
		out = append(out, *dmodel.NewMissingFieldErr(field + "." + basemodel.FieldId))
	}
	if item.ProductVariantId != nil && *item.ProductVariantId == "" {
		out = append(out, *dmodel.NewMissingFieldErr(field + "." + models.StockMoveFieldProductVariantId))
	}
	if item.DemandQuantity != nil {
		out = append(out, validateMoveQuantity(*item.DemandQuantity, field)...)
	}
	return out
}

// validateMoveQuantity refuses a line that could not move anything: a zero quantity would insert
// cleanly, reserve nothing and validate as complete.
func validateMoveQuantity(quantity decimal.Decimal, field string) ft.ClientErrors {
	if quantity.GreaterThan(decimal.Zero) {
		return nil
	}
	return ft.ClientErrors{*ft.NewBusinessViolation(
		field+"."+models.StockMoveFieldDemandQuantity,
		ft.ErrorKey("stock_move.quantity_must_be_positive", "inventory"),
		"a move must ask for a quantity greater than zero",
	)}
}

// refuseMoveListedTwice refuses a move both updated and removed, or named twice in one list: the
// second write would hit a row the first already changed.
func refuseMoveListedTwice(listed map[string]bool, moveId string, field string) ft.ClientErrors {
	if moveId == "" {
		return nil
	}
	if listed[moveId] {
		return ft.ClientErrors{*ft.NewBusinessViolation(field, ft.ErrorKey("stock_transfer.move_listed_twice", "inventory"),
			"move {{id}} is named more than once in this request", map[string]any{"id": moveId})}
	}
	listed[moveId] = true
	return nil
}

// applyMovesDelta expects the caller's transaction. Removals run first, then updates, then
// additions, so an addition's sequence follows every move the transfer keeps.
func applyMovesDelta(
	ctx corectx.Context, transferId string, delta itStock.TransferMovesDelta,
) (ft.ClientErrors, error) {
	writer, cErrs, err := newTransferMovesWriter(ctx, transferId)
	if err != nil || cErrs.Count() > 0 {
		return cErrs, err
	}
	if cErrs := writer.assertOwned(delta); cErrs.Count() > 0 {
		return cErrs, nil
	}
	if cErrs, err := writer.remove(ctx, delta.Remove); err != nil || cErrs.Count() > 0 {
		return cErrs, err
	}
	if cErrs, err := writer.update(ctx, delta.Update); err != nil || cErrs.Count() > 0 {
		return cErrs, err
	}
	return writer.add(ctx, delta.Add)
}

// transferMovesWriter applies a delta to one draft transfer's moves.
type transferMovesWriter struct {
	moveSvc      composable.CrudDomainService
	transfer     models.StockTransfer
	transferId   string
	owned        map[string]dmodel.DynamicFields
	nextSequence int
}

// newTransferMovesWriter reads the transfer and its moves, and refuses a transfer past draft: its
// moves are being reserved or were already executed.
func newTransferMovesWriter(
	ctx corectx.Context, transferId string,
) (*transferMovesWriter, ft.ClientErrors, error) {
	operation, err := loadTransferOperation(ctx, transferId)
	if err != nil {
		return nil, nil, err
	}
	if operation == nil {
		return nil, ft.ClientErrors{*ft.NewNotFoundError(basemodel.FieldId)}, nil
	}
	status := derefString(operation.Transfer.GetStatus())
	if status != models.StockTransferStatusDraft {
		return nil, ft.ClientErrors{*ft.NewBusinessViolation(itStock.TransferFieldMoves,
			ft.ErrorKey("stock_transfer.moves_not_editable", "inventory"),
			"moves can change only while the transfer is draft; this one is '{{status}}'",
			map[string]any{"status": status})}, nil
	}
	moveSvc, err := domainServiceFor(models.StockMoveSchemaName)
	if err != nil {
		return nil, nil, err
	}
	writer := &transferMovesWriter{
		moveSvc: moveSvc, transfer: operation.Transfer, transferId: transferId,
		owned: make(map[string]dmodel.DynamicFields, len(operation.Moves)), nextSequence: 1,
	}
	for _, move := range operation.Moves {
		writer.owned[readModelId(move, basemodel.FieldId)] = move
		if sequence := move.GetInt32(models.StockMoveFieldSequence); sequence != nil && int(*sequence) >= writer.nextSequence {
			writer.nextSequence = int(*sequence) + 1
		}
	}
	return writer, nil, nil
}

// assertOwned refuses a move of another transfer: the delta may only touch this transfer's moves.
func (this *transferMovesWriter) assertOwned(delta itStock.TransferMovesDelta) ft.ClientErrors {
	var out ft.ClientErrors
	for i, item := range delta.Update {
		out = append(out, this.refuseForeignMove(string(item.Id), valuesDeltaRowPath(movesFieldUpdate, i))...)
	}
	for i, moveId := range delta.Remove {
		out = append(out, this.refuseForeignMove(string(moveId), valuesDeltaRowPath(movesFieldRemove, i))...)
	}
	return out
}

func (this *transferMovesWriter) refuseForeignMove(moveId string, field string) ft.ClientErrors {
	if _, isOwned := this.owned[moveId]; isOwned {
		return nil
	}
	return ft.ClientErrors{*ft.NewBusinessViolation(field, ft.ErrorKey("stock_transfer.move_not_in_transfer", "inventory"),
		"move {{id}} does not belong to this transfer", map[string]any{"id": moveId})}
}

func (this *transferMovesWriter) remove(ctx corectx.Context, moveIds []string) (ft.ClientErrors, error) {
	for i, moveId := range moveIds {
		res, err := this.moveSvc.Delete(ctx, dmodel.DynamicFields{basemodel.FieldId: moveId})
		if err != nil {
			return nil, err
		}
		if res.ClientErrors.Count() > 0 {
			return prefixRowErrors(res.ClientErrors, valuesDeltaRowPath(movesFieldRemove, i)), nil
		}
	}
	return nil, nil
}

func (this *transferMovesWriter) update(
	ctx corectx.Context, items []itStock.TransferMoveUpdate,
) (ft.ClientErrors, error) {
	for i, item := range items {
		field := valuesDeltaRowPath(movesFieldUpdate, i)
		cmd := dmodel.DynamicFields{basemodel.FieldId: string(item.Id), basemodel.FieldEtag: item.Etag}
		if item.DemandQuantity != nil {
			cmd[models.StockMoveFieldDemandQuantity] = item.DemandQuantity.String()
			// Base equals demand while no UoM conversion exists; see newTransferMoveFields.
			cmd[models.StockMoveFieldBaseDemandQuantity] = item.DemandQuantity.String()
		}
		if item.UomId != nil {
			cmd[models.StockMoveFieldUomId] = nilIfEmpty(string(*item.UomId))
		}
		res, err := this.moveSvc.Update(ctx, cmd)
		if err != nil {
			return nil, err
		}
		if res.ClientErrors.Count() > 0 {
			return prefixRowErrors(res.ClientErrors, field), nil
		}
		if cErrs, err := this.replaceIfVariantChanged(ctx, item, field); err != nil || cErrs.Count() > 0 {
			return cErrs, err
		}
	}
	return nil, nil
}

// replaceIfVariantChanged swaps the move for a new one on the new variant, in the same position. A
// move's variant is fixed once created (its lines, reservations and valuation all name it), but a
// draft move has none of those yet, so replacing it loses nothing. The update above has already
// checked the etag and written the other changes.
func (this *transferMovesWriter) replaceIfVariantChanged(
	ctx corectx.Context, item itStock.TransferMoveUpdate, field string,
) (ft.ClientErrors, error) {
	current := this.owned[string(item.Id)]
	if item.ProductVariantId == nil ||
		string(*item.ProductVariantId) == readModelId(current, models.StockMoveFieldProductVariantId) {
		return nil, nil
	}
	res, err := this.moveSvc.Delete(ctx, dmodel.DynamicFields{basemodel.FieldId: string(item.Id)})
	if err != nil {
		return nil, err
	}
	if res.ClientErrors.Count() > 0 {
		return prefixRowErrors(res.ClientErrors, field), nil
	}
	line := transferMoveLine{variantId: string(*item.ProductVariantId), uomId: readModelId(current, models.StockMoveFieldUomId)}
	if item.UomId != nil {
		line.uomId = string(*item.UomId)
	}
	line.quantity = derefDecimal(current.GetDecimal(models.StockMoveFieldDemandQuantity))
	if item.DemandQuantity != nil {
		line.quantity = *item.DemandQuantity
	}
	sequence := current.GetInt32(models.StockMoveFieldSequence)
	if sequence == nil {
		return this.create(ctx, line, this.takeSequence(), field)
	}
	return this.create(ctx, line, int(*sequence), field)
}

func (this *transferMovesWriter) add(ctx corectx.Context, items []itStock.TransferMoveAdd) (ft.ClientErrors, error) {
	for i, item := range items {
		line := transferMoveLine{variantId: string(item.ProductVariantId), quantity: *item.DemandQuantity}
		if item.UomId != nil {
			line.uomId = string(*item.UomId)
		}
		if cErrs, err := this.create(ctx, line, this.takeSequence(), valuesDeltaRowPath(movesFieldAdd, i)); err != nil ||
			cErrs.Count() > 0 {
			return cErrs, err
		}
	}
	return nil, nil
}

func (this *transferMovesWriter) takeSequence() int {
	sequence := this.nextSequence
	this.nextSequence++
	return sequence
}

func (this *transferMovesWriter) create(
	ctx corectx.Context, line transferMoveLine, sequence int, field string,
) (ft.ClientErrors, error) {
	res, err := this.moveSvc.Create(ctx, newTransferMoveFields(this.transfer, this.transferId, sequence, line))
	if err != nil {
		return nil, err
	}
	if res.ClientErrors.Count() > 0 {
		return prefixRowErrors(res.ClientErrors, field), nil
	}
	return nil, nil
}

func readModelId(fields dmodel.DynamicFields, field string) string {
	if fields == nil {
		return ""
	}
	if id := fields.GetModelId(field); id != nil {
		return *id
	}
	return ""
}

func nilIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}
