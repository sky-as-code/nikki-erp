package services

import (
	"time"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// Physical inventory: entering a count, resetting it, and applying it as an adjustment. There is no
// adjustment header table; the count lives on the quant itself.
//
// All three bypass the quant engine's refusal of client writes, which is safe because none touches
// on_hand_quantity: enter and reset write only count metadata, and apply changes the balance solely
// through ApplyCorrectionMovement, keeping applyQuantDelta reachable from one place.

// EnterCount records what a physical count found, without changing the balance. It snapshots the
// current on-hand under the row lock into count_snapshot_quantity, which the staleness check at
// apply time depends on. Entering a count over a pending one is an allowed recount.
func (this *StockQuantDomainServiceImpl) EnterCount(
	ctx corectx.Context, quantId string, counted decimal.Decimal, reasonCode, reasonText string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	if vErrs := AssertCountEnterable(counted); vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	var result *dyn.OpResult[dyn.MutateResultData]
	err := withQuantTransaction(ctx, func(tranxCtx corectx.Context) error {
		locked, quant, vErrs, err := lockQuantById(tranxCtx, quantId)
		if err != nil || vErrs.Count() > 0 {
			result = clientErrorResult(vErrs)
			return err
		}

		fields := CountEntryFields(counted, locked.OnHand, reasonCode, reasonText)
		if err := updateQuantFields(tranxCtx, *quant, fields); err != nil {
			return err
		}
		result = mutateOk()
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// ResetCount abandons a pending count, leaving the balance untouched. It tolerates there being no
// pending count, so a double-click is not an error.
func (this *StockQuantDomainServiceImpl) ResetCount(
	ctx corectx.Context, quantId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	var result *dyn.OpResult[dyn.MutateResultData]
	err := withQuantTransaction(ctx, func(tranxCtx corectx.Context) error {
		_, quant, vErrs, err := lockQuantById(tranxCtx, quantId)
		if err != nil || vErrs.Count() > 0 {
			result = clientErrorResult(vErrs)
			return err
		}

		if err := updateQuantFields(tranxCtx, *quant, CountResetFields()); err != nil {
			return err
		}
		result = mutateOk()
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// ApplyAdjustment turns a pending count into a real balance change. The order matters: lock the
// quant, re-read on-hand inside that lock and compare with the snapshot (comparing a value read
// before the lock reproduces the race this prevents), generate the movement for a non-zero
// variance, then clear the count and stamp the history whether or not a movement ran.
func (this *StockQuantDomainServiceImpl) ApplyAdjustment(
	ctx corectx.Context, quantId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	var result *dyn.OpResult[dyn.MutateResultData]
	err := withQuantTransaction(ctx, func(tranxCtx corectx.Context) error {
		locked, quant, vErrs, err := lockQuantById(tranxCtx, quantId)
		if err != nil || vErrs.Count() > 0 {
			result = clientErrorResult(vErrs)
			return err
		}

		outcome, err := applyCountToQuant(tranxCtx, *quant, *locked)
		if err != nil {
			return err
		}
		result = outcome
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// ScheduleCount sets when a balance is next due to be counted. The worklist is a filtered search
// over next_count_date, so scheduling is just a field write. An empty date clears the schedule,
// taking the balance off the worklist without counting it.
func (this *StockQuantDomainServiceImpl) ScheduleCount(
	ctx corectx.Context, quantId string, nextCountDate string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	parsed, vErrs := parseOptionalDate(nextCountDate)
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	fields := map[string]any{models.StockQuantFieldNextCountDate: nil}
	if parsed != nil {
		fields[models.StockQuantFieldNextCountDate] = *parsed
	}
	return this.updateCountMetadata(ctx, quantId, fields)
}

// AssignCounter names who is responsible for counting a balance; an empty user id unassigns. The id
// is not checked against IAM, since inventory does not import it.
func (this *StockQuantDomainServiceImpl) AssignCounter(
	ctx corectx.Context, quantId string, userId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	value := any(userId)
	if userId == "" {
		value = nil
	}
	return this.updateCountMetadata(ctx, quantId, map[string]any{
		models.StockQuantFieldCountAssignedUser: value,
	})
}

// updateCountMetadata writes scheduling fields without a row lock: no balance arithmetic reads
// them, so a concurrent movement cannot invalidate them.
func (this *StockQuantDomainServiceImpl) updateCountMetadata(
	ctx corectx.Context, quantId string, fields map[string]any,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	engine, err := engineFor(models.StockQuantSchemaName)
	if err != nil {
		return nil, err
	}

	found, err := engine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{
		models.StockQuantFieldId: quantId,
	})
	if err != nil {
		return nil, errors.Wrap(err, "updateCountMetadata")
	}
	if found == nil || !found.HasData {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation(
			models.StockQuantSchemaName, "stock_quant.not_found",
			"no stock balance with id '"+quantId+"'"))
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	if err := updateQuantFields(ctx, *models.NewStockQuantFrom(found.Data), fields); err != nil {
		return nil, err
	}
	return mutateOk(), nil
}

// parseOptionalDate reads a yyyy-mm-dd date, treating empty as "clear it".
func parseOptionalDate(value string) (*time.Time, *ft.ClientErrors) {
	vErrs := ft.NewClientErrors()
	if value == "" {
		return nil, vErrs
	}

	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		vErrs.Append(*ft.NewBusinessViolation(
			models.StockQuantSchemaName, "stock_quant.next_count_date_malformed",
			"'next_count_date' must be a date in yyyy-mm-dd form"))
		return nil, vErrs
	}
	return &parsed, vErrs
}

// applyCountToQuant is the body of the apply, once the row is locked.
func applyCountToQuant(
	ctx corectx.Context, quant models.StockQuant, locked LockedQuant,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	if vErrs := AssertCountApplicable(derefBool(quant.GetCountQuantitySet())); vErrs.Count() > 0 {
		return clientErrorResult(vErrs), nil
	}

	snapshot := orZero(quant.GetCountSnapshotQuantity())
	if IsCountSnapshotStale(snapshot, locked.OnHand) {
		return clientErrorResult(StaleCountErrors(snapshot, locked.OnHand)), nil
	}

	variance := CountVariance(orZero(quant.GetCountedQuantity()), locked.OnHand)
	if !variance.IsZero() {
		vErrs, err := generateAdjustmentMovement(ctx, quant, locked, variance)
		if err != nil {
			return nil, err
		}
		if vErrs.Count() > 0 {
			return clientErrorResult(vErrs), nil
		}
	}

	// Reached on a zero variance too: the count is still resolved, or a balance confirmed correct
	// would stay permanently overdue on the worklist.
	if err := updateQuantFields(ctx, quant, CountAppliedFields(time.Now().UTC(), nil)); err != nil {
		return nil, err
	}
	return mutateOk(), nil
}

// generateAdjustmentMovement routes the variance through the shared correction helper. A positive
// variance flows from the inventory-loss location into the counted location; a negative variance
// flows the other way. Swapping them silently inverts the sign of the movement.
func generateAdjustmentMovement(
	ctx corectx.Context, quant models.StockQuant, locked LockedQuant, variance decimal.Decimal,
) (*ft.ClientErrors, error) {
	orgId := derefString(quant.GetOrgId())
	lossLocation, err := FindLocationByType(ctx, orgId, models.InventoryLocationUsageInventoryLoss)
	if err != nil {
		return nil, err
	}
	if lossLocation == nil {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation(
			models.InventoryLocationSchemaName,
			"stock_quant.no_inventory_loss_location",
			"this organisation has no inventory-loss location for an adjustment to balance against"))
		return vErrs, nil
	}

	countedLocation := derefString(quant.GetLocationId())
	request := CorrectionRequest{
		OrgId:                 orgId,
		ProductVariantId:      derefString(quant.GetProductVariantId()),
		Quantity:              variance.Abs(),
		SourceLocationId:      derefString(lossLocation.GetId()),
		DestinationLocationId: countedLocation,
		LotRef:                locked.LotRef,
		PackageRef:            locked.PackageRef,
		OwnerRef:              locked.OwnerRef,
		OriginReference:       "count:" + derefString(quant.GetId()),
		IsInventoryAdjustment: true,
	}
	if variance.IsNegative() {
		request.SourceLocationId = countedLocation
		request.DestinationLocationId = derefString(lossLocation.GetId())
	}

	_, vErrs, err := ApplyCorrectionMovement(ctx, request)
	return vErrs, err
}

// lockQuantById reads a quant and takes the row lock its dimension belongs to. The lock is keyed by
// (org, variant, location), the unit a movement contends over; the matching row is then picked out
// of the locked set by id.
func lockQuantById(
	ctx corectx.Context, quantId string,
) (*LockedQuant, *models.StockQuant, *ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()

	engine, err := engineFor(models.StockQuantSchemaName)
	if err != nil {
		return nil, nil, vErrs, err
	}

	found, err := engine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{
		models.StockQuantFieldId: quantId,
	})
	if err != nil {
		return nil, nil, vErrs, errors.Wrap(err, "lockQuantById")
	}
	if found == nil || !found.HasData {
		vErrs.Append(*ft.NewBusinessViolation(
			models.StockQuantSchemaName, "stock_quant.not_found",
			"no stock balance with id '"+quantId+"'"))
		return nil, nil, vErrs, nil
	}

	quant := models.NewStockQuantFrom(found.Data)
	locked, err := lockQuantRow(ctx, engine, *quant, quantId)
	if err != nil {
		return nil, nil, vErrs, err
	}
	if locked == nil {
		vErrs.Append(*ft.NewBusinessViolation(
			models.StockQuantSchemaName, "stock_quant.not_found",
			"stock balance '"+quantId+"' disappeared while being locked"))
		return nil, nil, vErrs, nil
	}
	return locked, quant, vErrs, nil
}

func lockQuantRow(
	ctx corectx.Context, engine drif.DynamicResourceEngine, quant models.StockQuant, quantId string,
) (*LockedQuant, error) {
	rows, err := LockQuantsForUpdate(ctx, engine.ResourceRepository().GetBaseRepo(), QuantLockKey{
		OrgId:            model.Id(derefString(quant.GetOrgId())),
		ProductVariantId: model.Id(derefString(quant.GetProductVariantId())),
		LocationId:       model.Id(derefString(quant.GetLocationId())),
	})
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if string(row.Id) == quantId {
			return &row, nil
		}
	}
	return nil, nil
}

// updateQuantFields writes count metadata straight through the repository, bypassing the quant
// engine's blanket refusal of client writes. That is safe only because none of these fields is
// on_hand_quantity or reserved_quantity.
func updateQuantFields(
	ctx corectx.Context, quant models.StockQuant, fields map[string]any,
) error {
	engine, err := engineFor(models.StockQuantSchemaName)
	if err != nil {
		return err
	}

	update := dmodel.DynamicFields{
		models.StockQuantFieldId: derefString(quant.GetId()),
		basemodel.FieldEtag:      derefString(quant.GetEtag()),
	}
	for key, value := range fields {
		update[key] = value
	}

	_, err = engine.ResourceRepository().Update(ctx, update)
	return errors.Wrap(err, "updateQuantFields")
}

// withQuantTransaction runs body inside one transaction on a cloned context, so the transaction is
// not left visible to whatever runs next. There is no join-an-existing branch because BeginTx
// returns ErrTxNested.
func withQuantTransaction(ctx corectx.Context, body func(tranxCtx corectx.Context) error) error {
	engine, err := engineFor(models.StockQuantSchemaName)
	if err != nil {
		return err
	}

	tranx, err := engine.ResourceRepository().BeginTransaction(ctx)
	if err != nil {
		return errors.Wrap(err, "withQuantTransaction")
	}
	defer tranx.Rollback()

	tranxCtx := corectx.CloneRequestContext(ctx)
	tranxCtx.SetDbTranx(tranx)

	if err := body(tranxCtx); err != nil {
		return err
	}
	return errors.Wrap(tranx.Commit(), "withQuantTransaction")
}

// clientErrorResult wraps validation failures in the mutation envelope, or reports success when
// there are none.
func clientErrorResult(vErrs *ft.ClientErrors) *dyn.OpResult[dyn.MutateResultData] {
	if vErrs == nil || vErrs.Count() == 0 {
		return mutateOk()
	}
	return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}
}
