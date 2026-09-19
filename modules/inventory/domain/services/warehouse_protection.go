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

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// Protecting committed stock (CR-INV-SALES-WH-RESERVATION §4.5).
//
// Every path that lowers a warehouse's eligible on-hand, or raises what is claimed on it, runs
// under the scope's guard and leaves H >= R behind. The check is the same everywhere: after the
// write, the scope's availability must not be negative. Consumption passes it because H and R
// fall together; a scrap, a downward count or a foreign issue that would eat into a commitment
// does not, and is refused rather than applied with stock that is no longer there.
//
// The refusal travels as a Go error so the enclosing transaction rolls back the write it refused;
// the service boundary turns it into a client error with ProtectionViolationOf.

const ReasonProtectionViolation = "stock_reservation.protection_violation"

// ProtectionViolation is the refusal, carried as an error until the transaction has been undone.
type ProtectionViolation struct {
	Errors *ft.ClientErrors
}

func (this *ProtectionViolation) Error() string {
	return "the operation would leave a warehouse with less stock than it has committed"
}

// ProtectionViolationOf unwraps the refusal at a service boundary, after the rollback.
func ProtectionViolationOf(err error) (*ft.ClientErrors, bool) {
	var violation *ProtectionViolation
	if errors.As(err, &violation) {
		return violation.Errors, true
	}
	return nil, false
}

// protectionResultOf answers a mutation with the refusal, or nil when err is something else.
func protectionResultOf(err error) *dyn.OpResult[dyn.MutateResultData] {
	if vErrs, ok := ProtectionViolationOf(err); ok {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}
	}
	return nil
}

// warehouseScopeOfLocation resolves the scope a movement at a location falls under. A location
// outside any warehouse (a supplier, a customer, transit) has no scope and needs no guard.
func warehouseScopeOfLocation(
	ctx corectx.Context, orgId, variantId, locationId string,
) (GuardKey, bool, error) {
	if locationId == "" || variantId == "" {
		return GuardKey{}, false, nil
	}
	location, err := findRecord(ctx, models.InventoryLocationSchemaName, models.InventoryLocationFieldId, locationId)
	if err != nil {
		return GuardKey{}, false, err
	}
	warehouseId := stringOf(location, models.InventoryLocationFieldWarehouseId)
	if warehouseId == "" {
		return GuardKey{}, false, nil
	}
	if orgId == "" {
		orgId = stringOf(location, models.InventoryLocationFieldOrgId)
	}
	return GuardKey{
		OrgId:            model.Id(orgId),
		WarehouseId:      model.Id(warehouseId),
		ProductVariantId: model.Id(variantId),
	}, true, nil
}

// lockWarehouseScopes takes the guards of every scope, in lock order, and returns the database
// clock read once they are held. Quant locks come after this, never before.
func lockWarehouseScopes(ctx corectx.Context, keys []GuardKey) (time.Time, error) {
	guardRepo, err := repoFor(models.WarehouseProductGuardSchemaName)
	if err != nil {
		return time.Time{}, err
	}
	if _, err := LockGuardsForUpdate(ctx, guardRepo.GetBaseRepo(), keys); err != nil {
		return time.Time{}, err
	}
	return DbNowUnderLock(ctx, guardRepo.GetBaseRepo())
}

// assertNoBackingShortfall recomputes the scope under its held guard and refuses when the
// stock no longer covers what is committed on it.
func assertNoBackingShortfall(ctx corectx.Context, key GuardKey, now time.Time) (*ft.ClientErrors, error) {
	availability, err := ComputeWarehouseAvailability(ctx, key, now)
	if err != nil {
		return nil, err
	}
	vErrs := ft.NewClientErrors()
	if availability.Available.IsNegative() {
		vErrs.Append(*ft.NewBusinessViolation(models.StockReservationSchemaName, ReasonProtectionViolation,
			"warehouse '"+string(key.WarehouseId)+"' would hold "+availability.EligibleOnHand.String()+
				" of product variant '"+string(key.ProductVariantId)+"' against "+availability.EffectiveReserved.String()+
				" committed; the operation is refused",
			map[string]any{
				"warehouse_id":       string(key.WarehouseId),
				"product_variant_id": string(key.ProductVariantId),
				"eligible_on_hand":   availability.EligibleOnHand.String(),
				"effective_reserved": availability.EffectiveReserved.String(),
			}))
	}
	return vErrs, nil
}

// warehouseAvailableAt is the scope's A for a caller deciding how much it may still take.
func warehouseAvailableAt(ctx corectx.Context, key GuardKey, now time.Time) (decimal.Decimal, error) {
	availability, err := ComputeWarehouseAvailability(ctx, key, now)
	if err != nil {
		return decimal.Zero, err
	}
	return availability.Available, nil
}

// capToWarehouseAvailability bounds a claim by the scope's availability, never below zero.
func capToWarehouseAvailability(outstanding, available decimal.Decimal) decimal.Decimal {
	if available.IsNegative() {
		return decimal.Zero
	}
	if outstanding.GreaterThan(available) {
		return available
	}
	return outstanding
}

// assertLocationNotBackingReservations refuses to take a location out of its warehouse's sellable
// stock (suspend, archive, move elsewhere) while the warehouse needs that stock to cover its
// commitments. The stock is not touched by such a change, but it stops being eligible, which is
// the same thing to a reservation.
func assertLocationNotBackingReservations(
	ctx corectx.Context, location models.InventoryLocation,
) (*ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()
	warehouseId := derefId(location.GetWarehouseId())
	if warehouseId == "" || derefString(location.GetLocationUsage()) != models.InventoryLocationUsageInternal {
		return vErrs, nil
	}

	engine, err := repoFor(models.StockQuantSchemaName)
	if err != nil {
		return vErrs, err
	}
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.StockQuantFieldLocationId, dmodel.Equals, derefId(location.GetId())),
		*dmodel.NewSearchNode().NewCondition(models.StockQuantFieldOwnerRef, dmodel.Equals, ""),
	)
	onHandHere := map[string]decimal.Decimal{}
	err = scanAllRows(ctx, engine, graph, func(row dmodel.DynamicFields) {
		quant := models.NewStockQuantFrom(row)
		variantId := derefId(quant.GetProductVariantId())
		onHandHere[variantId] = onHandHere[variantId].Add(derefDecimal(quant.GetOnHandQuantity()))
	})
	if err != nil {
		return vErrs, errors.Wrap(err, "assertLocationNotBackingReservations")
	}

	now := time.Now().UTC()
	for variantId, here := range onHandHere {
		if !here.IsPositive() {
			continue
		}
		availability, err := ComputeWarehouseAvailability(ctx, GuardKey{
			OrgId:            model.Id(derefId(location.GetOrgId())),
			WarehouseId:      model.Id(warehouseId),
			ProductVariantId: model.Id(variantId),
		}, now)
		if err != nil {
			return vErrs, err
		}
		if availability.Available.Sub(here).IsNegative() {
			vErrs.Append(*ft.NewBusinessViolation(models.InventoryLocationSchemaName,
				"inventory_location.backs_reservations",
				"the location holds stock of product variant '"+variantId+"' that warehouse '"+warehouseId+
					"' needs to cover its reservations; release or consume them first"))
		}
	}
	return vErrs, nil
}

// assertWarehouseHasNoEffectiveReservations refuses to suspend or archive a warehouse while a
// reservation is still in force on it: the commitment would silently lose its backing.
func assertWarehouseHasNoEffectiveReservations(
	ctx corectx.Context, orgId, warehouseId string,
) (*ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()
	engine, err := repoFor(models.StockReservationSchemaName)
	if err != nil {
		return vErrs, err
	}
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.StockReservationFieldOrgId, dmodel.Equals, orgId),
		*dmodel.NewSearchNode().NewCondition(models.StockReservationFieldWarehouseId, dmodel.Equals, warehouseId),
		*dmodel.NewSearchNode().NewCondition(models.StockReservationFieldStatus, dmodel.Equals, models.StockReservationStatusActive),
	)
	now := time.Now().UTC()
	inForce := 0
	err = scanAllRows(ctx, engine, graph, func(row dmodel.DynamicFields) {
		if IsReservationEffective(*models.NewStockReservationFrom(row), now) {
			inForce++
		}
	})
	if err != nil {
		return vErrs, errors.Wrap(err, "assertWarehouseHasNoEffectiveReservations")
	}
	if inForce > 0 {
		vErrs.Append(*ft.NewBusinessViolation(models.WarehouseSchemaName, "warehouse.has_active_reservations",
			"the warehouse still holds stock for reservations in force; release or consume them first"))
	}
	return vErrs, nil
}
