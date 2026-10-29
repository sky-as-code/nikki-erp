package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// What the application layer needs from the warehouse domain, exported narrowly: the app package
// orchestrates two domain services across one transaction, and the rules stay in this package.

// IncomingFlowLocations lists the stops goods make on the way in, beyond the root itself.
func IncomingFlowLocations(flow string) []SystemLocationSpec {
	return toExportedSpecs(incomingFlowLocations(flow))
}

// OutgoingFlowLocations lists the stops goods make on the way out, beyond the root itself.
func OutgoingFlowLocations(flow string) []SystemLocationSpec {
	return toExportedSpecs(outgoingFlowLocations(flow))
}

// ObsoleteSystemLocations lists the stops a flow change leaves unused, so the caller can suspend
// them. Never deleted: the movement history that passed through them still names them.
func ObsoleteSystemLocations(previousFlow string, nextFlow string, outgoing bool) []SystemLocationSpec {
	return toExportedSpecs(obsoleteSystemLocations(previousFlow, nextFlow, outgoing))
}

// SystemLocationSpec describes one stop of a warehouse flow: a location the warehouse creates for
// itself under its root, told apart from the root's other children by its purpose.
type SystemLocationSpec struct {
	Name    string
	Purpose string
}

func toExportedSpecs(specs []systemLocationSpec) []SystemLocationSpec {
	exported := make([]SystemLocationSpec, 0, len(specs))
	for _, spec := range specs {
		exported = append(exported, SystemLocationSpec(spec))
	}
	return exported
}

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

// WriteLocationStatus sets a location's operational state without re-running the lifecycle rules
// the caller has already applied.
func WriteLocationStatus(ctx corectx.Context, locationId string, status string) error {
	return writeLocationFieldsDirect(ctx, locationId, dmodel.DynamicFields{
		models.InventoryLocationFieldStatus: status,
	})
}

// MarkLocationSystemGenerated flags a location as one the warehouse created for itself. Written
// here rather than through Create, which strips the flag from anything a caller supplies to stop a
// client minting a location that then refuses to be archived.
func MarkLocationSystemGenerated(ctx corectx.Context, locationId string) error {
	return writeLocationFieldsDirect(ctx, locationId, dmodel.DynamicFields{
		models.InventoryLocationFieldIsSystemGenerated: true,
	})
}

// FindWarehouseFlowLocation resolves one stop of a warehouse flow: the system-generated child of
// the warehouse's root with the stop's purpose. Nil when the warehouse has no such stop.
func FindWarehouseFlowLocation(
	ctx corectx.Context, warehouse models.Warehouse, purpose string,
) (*models.InventoryLocation, error) {
	rootId := derefId(warehouse.GetRootLocationId())
	if rootId == "" {
		return nil, nil
	}
	engine, err := repoFor(models.InventoryLocationSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.InventoryLocationFieldParentLocationId, dmodel.Equals, rootId),
		*dmodel.NewSearchNode().NewCondition(models.InventoryLocationFieldIsSystemGenerated, dmodel.Equals, true),
		*dmodel.NewSearchNode().NewCondition(models.InventoryLocationFieldPurpose, dmodel.Equals, purpose),
	)
	found, err := engine.Search(ctx, dyn.RepoSearchParam{Graph: graph, Page: 0, Size: 1})
	if err != nil {
		return nil, errors.Wrap(err, "FindWarehouseFlowLocation")
	}
	if found == nil || !found.HasData || len(found.Data.Items) == 0 {
		return nil, nil
	}
	return models.NewInventoryLocationFrom(found.Data.Items[0]), nil
}

// ResolveWarehouseStop turns a stop named in a flow leg into its location id: "Stock" is the
// warehouse's root, any other stop its flow location. Empty when the warehouse lacks the stop.
func ResolveWarehouseStop(ctx corectx.Context, warehouse models.Warehouse, stopName string) (string, error) {
	if stopName == warehouseStockLocationCode {
		return derefId(warehouse.GetRootLocationId()), nil
	}
	purpose := stopPurpose(stopName)
	if purpose == "" {
		return "", nil
	}
	found, err := FindWarehouseFlowLocation(ctx, warehouse, purpose)
	if err != nil || found == nil {
		return "", err
	}
	return derefString(found.GetId()), nil
}

// stopPurpose looks a stop up in the three-step flows, which name every stop there is.
func stopPurpose(stopName string) string {
	allStops := append(incomingFlowLocations(models.WarehouseFlowThreeStep),
		outgoingFlowLocations(models.WarehouseFlowThreeStep)...)
	for _, spec := range allStops {
		if spec.Name == stopName {
			return spec.Purpose
		}
	}
	return ""
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
