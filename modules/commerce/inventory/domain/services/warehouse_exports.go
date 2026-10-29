package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// What the application layer needs from the warehouse domain, exported narrowly: the app package
// orchestrates two domain services across one transaction, and the rules stay in this package.

// IsKnownFlow reports whether a value is one of the three flow settings.
func IsKnownFlow(flow string) bool {
	return isKnownFlow(flow)
}

// LoadWarehouse fetches one warehouse for the application layer.
func LoadWarehouse(
	ctx corectx.Context, warehouseId string,
) (*models.Warehouse, *ft.ClientErrors, error) {
	return loadWarehouseById(ctx, warehouseId)
}

// WriteWarehouseFlow stores a new flow setting and nothing else: provisioning the locations the
// flow needs is the caller's next step, inside the same transaction.
func WriteWarehouseFlow(
	ctx corectx.Context, warehouseId string, field string, flow string,
) error {
	return writeWarehouseFields(ctx, warehouseId, dmodel.DynamicFields{field: flow})
}

// MarkLocationSystemGenerated flags a location as one the warehouse created for itself. Written
// here rather than through Create, which strips the flag from anything a caller supplies to stop a
// client minting a location that then refuses to be archived.
func MarkLocationSystemGenerated(ctx corectx.Context, locationId string) error {
	return writeLocationFieldsDirect(ctx, locationId, dmodel.DynamicFields{
		models.InventoryLocationFieldIsSystemGenerated: true,
	})
}

// ResolveWarehouseStop turns a stop named in a flow leg into its location id, read off the
// warehouse FK that holds it: "Stock" is the stock location, any other stop its flow location.
// Empty when the warehouse lacks the stop.
func ResolveWarehouseStop(warehouse models.Warehouse, stopName string) string {
	field := stopField(stopName)
	if field == "" {
		return ""
	}
	return derefId(warehouse.GetFlowLocationId(field))
}

// WithWarehouseTransaction runs body inside one transaction on a cloned context. Warehouse creation
// and flow reconfiguration each write a warehouse and its locations, and half of either result is
// useless.
func WithWarehouseTransaction(ctx corectx.Context, body func(tranxCtx corectx.Context) error) error {
	engine, err := repoFor(models.WarehouseSchemaName)
	if err != nil {
		return err
	}

	tranx, err := engine.BeginTransaction(ctx)
	if err != nil {
		return errors.Wrap(err, "WithWarehouseTransaction")
	}
	defer tranx.Rollback()

	tranxCtx := corectx.CloneRequestContext(ctx)
	tranxCtx.SetDbTranx(tranx)

	if err := body(tranxCtx); err != nil {
		return err
	}
	return errors.Wrap(tranx.Commit(), "WithWarehouseTransaction")
}

// writeLocationFieldsDirect writes location fields through the repository, bypassing the service
// overrides. Only for callers that have already validated the change.
func writeLocationFieldsDirect(
	ctx corectx.Context, locationId string, fields dmodel.DynamicFields,
) error {
	engine, err := repoFor(models.InventoryLocationSchemaName)
	if err != nil {
		return err
	}

	update := dmodel.DynamicFields{models.InventoryLocationFieldId: locationId}
	for key, value := range fields {
		update[key] = value
	}
	_, err = engine.Update(ctx, update)
	return errors.Wrap(err, "writeLocationFieldsDirect")
}
