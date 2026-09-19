package services

import (
	"time"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/array"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// Warehouse-level availability, the figure every warehouse reservation is decided on:
//
//	H  eligible on-hand: quants at the warehouse's own internal, active, unarchived locations,
//	   owned by the company, while the warehouse itself is active and unarchived.
//	R  effective commitment: the remainder of every warehouse reservation still in force, plus
//	   the quant-level reserved figure of the same quants (the legacy transfer-based holds).
//	A  H - R.
//
// The two legs of R never overlap: a warehouse reservation writes nothing to a quant, and a
// transfer hold writes only the quant, so each committed unit is counted exactly once.
//
// A parent warehouse never includes a child's stock: only locations whose warehouse_id is the
// requested warehouse count. Expiry is decided by the clock the caller passes, read from the
// database after the guard lock was taken, never by a stored status.

// WarehouseAvailability is the answer for one (org, warehouse, variant) scope at AsOf.
type WarehouseAvailability struct {
	EligibleOnHand    decimal.Decimal
	EffectiveReserved decimal.Decimal
	Available         decimal.Decimal
	AsOf              time.Time
}

// Shortage is how much of a request cannot be met, never negative.
func (this WarehouseAvailability) Shortage(requested decimal.Decimal) decimal.Decimal {
	shortage := requested.Sub(this.Available)
	if shortage.IsNegative() {
		return decimal.Zero
	}
	return shortage
}

// IsReservationEffective is the one definition of "still in force", shared by availability,
// consume, release, protect and the expiry sweep. The interval is half-open: a reservation whose
// deadline equals now has already lapsed.
func IsReservationEffective(reservation models.StockReservation, now time.Time) bool {
	if derefString(reservation.GetStatus()) != models.StockReservationStatusActive {
		return false
	}
	until := reservation.GetReservedUntil()
	return until == nil || until.GoTime().After(now)
}

// EffectiveReservedQuantity is what a reservation commits at now: its remainder while in force,
// zero otherwise.
func EffectiveReservedQuantity(reservation models.StockReservation, now time.Time) decimal.Decimal {
	if !IsReservationEffective(reservation, now) {
		return decimal.Zero
	}
	return reservation.RemainingQuantity()
}

// ComputeWarehouseAvailability reads H, R and A for one scope. It writes nothing and takes no lock;
// a caller that will act on the answer must already hold the scope's guard and pass the database
// clock it read after locking, because a figure read before the lock is stale by definition.
func ComputeWarehouseAvailability(
	ctx corectx.Context, key GuardKey, now time.Time,
) (WarehouseAvailability, error) {
	result := WarehouseAvailability{AsOf: now}

	usable, err := isWarehouseUsable(ctx, string(key.WarehouseId))
	if err != nil {
		return result, err
	}
	if !usable {
		return finishAvailability(result), nil
	}

	locationIds, err := eligibleLocationIds(ctx, key)
	if err != nil {
		return result, err
	}

	if len(locationIds) > 0 {
		onHand, legacyReserved, err := sumEligibleQuants(ctx, key, locationIds)
		if err != nil {
			return result, err
		}
		result.EligibleOnHand = onHand
		result.EffectiveReserved = legacyReserved
	}

	reserved, err := sumEffectiveReservations(ctx, key, now)
	if err != nil {
		return result, err
	}
	result.EffectiveReserved = result.EffectiveReserved.Add(reserved)

	return finishAvailability(result), nil
}

func finishAvailability(result WarehouseAvailability) WarehouseAvailability {
	result.Available = result.EligibleOnHand.Sub(result.EffectiveReserved)
	return result
}

// isWarehouseUsable reports whether a warehouse is active and unarchived, the state in which its
// stock may be committed. The location lifecycle uses the same rule for a location's owner.
func isWarehouseUsable(ctx corectx.Context, warehouseId string) (bool, error) {
	if warehouseId == "" {
		return false, nil
	}

	engine, err := repoFor(models.WarehouseSchemaName)
	if err != nil {
		return false, err
	}
	found, err := engine.FindByKeys(ctx, dmodel.DynamicFields{
		models.WarehouseFieldId: warehouseId,
	})
	if err != nil {
		return false, errors.Wrap(err, "isWarehouseUsable")
	}
	if found == nil || !found.HasData {
		return false, nil
	}

	warehouse := models.NewWarehouseFrom(found.Data)
	archived := derefBool(warehouse.GetIsArchived())
	return !archived && derefString(warehouse.GetStatus()) == models.WarehouseStatusActive, nil
}

// eligibleLocationIds lists the locations whose stock is sellable from the warehouse: its
// internal, active, unarchived locations, and their descendants. Quarantine, scrap, loss, transit
// and customer or vendor locations are excluded by usage; a location moved to another warehouse
// drops out by ownership.
//
// Descendants are walked because a child may leave warehouse_id NULL and inherit the owner from
// its parent - a kiosk slot is provisioned that way. Matching warehouse_id alone would make the
// stock in those slots invisible to availability while it is plainly in the warehouse, so an
// order would be refused for want of goods sitting in the machine that owes them.
//
// This is the single place a future eligibility rule (a quarantine purpose, say) plugs in.
func eligibleLocationIds(ctx corectx.Context, key GuardKey) ([]string, error) {
	engine, err := repoFor(models.InventoryLocationSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.InventoryLocationFieldOrgId, dmodel.Equals, string(key.OrgId)),
		*dmodel.NewSearchNode().NewCondition(models.InventoryLocationFieldWarehouseId, dmodel.Equals, string(key.WarehouseId)),
		*dmodel.NewSearchNode().NewCondition(models.InventoryLocationFieldLocationUsage, dmodel.Equals, models.InventoryLocationUsageInternal),
		*dmodel.NewSearchNode().NewCondition(models.InventoryLocationFieldStatus, dmodel.Equals, models.InventoryLocationStatusActive),
		*dmodel.NewSearchNode().NewCondition(basemodel.FieldIsArchived, dmodel.Equals, false),
	)

	ids := make([]string, 0, 16)
	err = scanAllRows(ctx, engine, graph, func(row dmodel.DynamicFields) {
		ids = append(ids, derefId(models.NewInventoryLocationFrom(row).GetId()))
	})
	if err != nil {
		return nil, errors.Wrap(err, "eligibleLocationIds")
	}

	ids, err = appendUsableDescendants(ctx, key, ids)
	return ids, errors.Wrap(err, "eligibleLocationIds")
}

// appendUsableDescendants extends the owned locations with every internal, active, unarchived
// location beneath them whose own warehouse_id is unset. A child that names a warehouse is left
// to that warehouse: the schema requires it to be the same one, and honouring the explicit value
// keeps the "moved to another warehouse drops out" rule above intact.
//
// The walk is breadth-first over parent_location_id and stops when a level adds nothing, so a
// tree of any depth costs one query per level. Cycles cannot be introduced by the engine, and the
// seen set makes one harmless here anyway.
func appendUsableDescendants(
	ctx corectx.Context, key GuardKey, owned []string,
) ([]string, error) {
	if len(owned) == 0 {
		return owned, nil
	}

	engine, err := repoFor(models.InventoryLocationSchemaName)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool, len(owned)*2)
	for _, id := range owned {
		seen[id] = true
	}

	all := owned
	frontier := owned
	for len(frontier) > 0 {
		graph := &dmodel.SearchGraph{}
		graph.And(
			*dmodel.NewSearchNode().NewCondition(models.InventoryLocationFieldOrgId, dmodel.Equals, string(key.OrgId)),
			*dmodel.NewSearchNode().NewCondition(models.InventoryLocationFieldParentLocationId, dmodel.In,
				array.Map(frontier, func(id string) any { return id })...),
			*dmodel.NewSearchNode().NewCondition(models.InventoryLocationFieldLocationUsage, dmodel.Equals, models.InventoryLocationUsageInternal),
			*dmodel.NewSearchNode().NewCondition(models.InventoryLocationFieldStatus, dmodel.Equals, models.InventoryLocationStatusActive),
			*dmodel.NewSearchNode().NewCondition(basemodel.FieldIsArchived, dmodel.Equals, false),
		)

		next := make([]string, 0, 16)
		err = scanAllRows(ctx, engine, graph, func(row dmodel.DynamicFields) {
			location := models.NewInventoryLocationFrom(row)
			id := derefId(location.GetId())
			if id == "" || seen[id] {
				return
			}
			seen[id] = true
			// A child naming its own warehouse belongs to that one, by the schema's rule that a
			// location and its parent must agree; only the inheriting children are ours to add.
			if derefId(location.GetWarehouseId()) == "" {
				all = append(all, id)
			}
			next = append(next, id)
		})
		if err != nil {
			return nil, err
		}
		frontier = next
	}
	return all, nil
}

// sumEligibleQuants totals on-hand and quant-level reserved over the company-owned quants of the
// variant at the given locations. Consigned stock (a non-empty owner_ref) is not the company's to
// sell and is left out of both figures.
func sumEligibleQuants(
	ctx corectx.Context, key GuardKey, locationIds []string,
) (decimal.Decimal, decimal.Decimal, error) {
	engine, err := repoFor(models.StockQuantSchemaName)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.StockQuantFieldOrgId, dmodel.Equals, string(key.OrgId)),
		*dmodel.NewSearchNode().NewCondition(models.StockQuantFieldProductVariantId, dmodel.Equals, string(key.ProductVariantId)),
		*dmodel.NewSearchNode().NewCondition(models.StockQuantFieldLocationId, dmodel.In, toAnySlice(locationIds)...),
		*dmodel.NewSearchNode().NewCondition(models.StockQuantFieldOwnerRef, dmodel.Equals, ""),
	)

	onHand, reserved := decimal.Zero, decimal.Zero
	err = scanAllRows(ctx, engine, graph, func(row dmodel.DynamicFields) {
		quant := models.NewStockQuantFrom(row)
		onHand = onHand.Add(derefDecimal(quant.GetOnHandQuantity()))
		reserved = reserved.Add(derefDecimal(quant.GetReservedQuantity()))
	})
	return onHand, reserved, errors.Wrap(err, "sumEligibleQuants")
}

// sumEffectiveReservations totals the remainder of every warehouse reservation of the scope that
// is still in force at now. Rows are filtered on the stored status only and the deadline is
// applied in Go against the caller's clock, so a lapsed row never counts however stale it is.
func sumEffectiveReservations(
	ctx corectx.Context, key GuardKey, now time.Time,
) (decimal.Decimal, error) {
	reservations, err := activeReservationsOfScope(ctx, key)
	if err != nil {
		return decimal.Zero, err
	}
	total := decimal.Zero
	for _, reservation := range reservations {
		total = total.Add(EffectiveReservedQuantity(reservation, now))
	}
	return total, nil
}

// activeReservationsOfScope lists the stored-active reservations of one scope. Some may have
// lapsed; the caller applies IsReservationEffective with its own clock.
func activeReservationsOfScope(ctx corectx.Context, key GuardKey) ([]models.StockReservation, error) {
	engine, err := repoFor(models.StockReservationSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.StockReservationFieldOrgId, dmodel.Equals, string(key.OrgId)),
		*dmodel.NewSearchNode().NewCondition(models.StockReservationFieldWarehouseId, dmodel.Equals, string(key.WarehouseId)),
		*dmodel.NewSearchNode().NewCondition(models.StockReservationFieldProductVariantId, dmodel.Equals, string(key.ProductVariantId)),
		*dmodel.NewSearchNode().NewCondition(models.StockReservationFieldStatus, dmodel.Equals, models.StockReservationStatusActive),
	)

	reservations := make([]models.StockReservation, 0, 8)
	err = scanAllRows(ctx, engine, graph, func(row dmodel.DynamicFields) {
		reservations = append(reservations, *models.NewStockReservationFrom(row))
	})
	return reservations, errors.Wrap(err, "activeReservationsOfScope")
}

// scanAllRows pages through every row matching graph, reading through to the last page rather
// than truncating: a partial total would under-count a commitment, which is the one error an
// availability figure must never make.
func scanAllRows(
	ctx corectx.Context, engine composable.CrudRepository, graph *dmodel.SearchGraph,
	visit func(row dmodel.DynamicFields),
) error {
	for page := 0; ; page++ {
		found, err := engine.Search(ctx, dyn.RepoSearchParam{
			Graph: graph,
			Page:  page,
			Size:  usageScanPageSize,
		})
		if err != nil {
			return err
		}
		if found == nil || !found.HasData || len(found.Data.Items) == 0 {
			return nil
		}
		for _, row := range found.Data.Items {
			visit(row)
		}
		if len(found.Data.Items) < usageScanPageSize {
			return nil
		}
	}
}

// guardKeyOf builds the scope key of a reservation row.
func guardKeyOf(reservation models.StockReservation) GuardKey {
	return GuardKey{
		OrgId:            model.Id(derefId(reservation.GetOrgId())),
		WarehouseId:      model.Id(derefId(reservation.GetWarehouseId())),
		ProductVariantId: model.Id(derefId(reservation.GetProductVariantId())),
	}
}
