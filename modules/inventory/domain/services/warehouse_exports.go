package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// What the application layer needs from the warehouse domain, exported narrowly: the app package
// orchestrates two domain services across one transaction, and the rules stay in this package.

// RequiredSystemLocations lists the locations a warehouse must have for its two flows.
func RequiredSystemLocations(incomingFlow string, outgoingFlow string) []SystemLocationSpec {
	return toExportedSpecs(requiredSystemLocations(incomingFlow, outgoingFlow))
}

// IncomingFlowLocations lists the stops goods make on the way in, beyond Stock itself.
func IncomingFlowLocations(flow string) []SystemLocationSpec {
	return toExportedSpecs(incomingFlowLocations(flow))
}

// OutgoingFlowLocations lists the stops goods make on the way out, beyond Stock itself.
func OutgoingFlowLocations(flow string) []SystemLocationSpec {
	return toExportedSpecs(outgoingFlowLocations(flow))
}

// ObsoleteSystemLocations lists the stops a flow change leaves unused, so the caller can suspend
// them. Never deleted: the movement history that passed through them still names them.
func ObsoleteSystemLocations(previousFlow string, nextFlow string, outgoing bool) []SystemLocationSpec {
	return toExportedSpecs(obsoleteSystemLocations(previousFlow, nextFlow, outgoing))
}

// SystemLocationSpec describes one location a warehouse creates for itself.
type SystemLocationSpec struct {
	Code    string
	Purpose string
}

func toExportedSpecs(specs []systemLocationSpec) []SystemLocationSpec {
	exported := make([]SystemLocationSpec, 0, len(specs))
	for _, spec := range specs {
		exported = append(exported, SystemLocationSpec{Code: spec.Code, Purpose: spec.Purpose})
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

// FindWarehouseLocationByCode resolves one of a warehouse's own locations by its code. Codes are
// unique within an org and system locations carry the codes the flow topology names.
func FindWarehouseLocationByCode(
	ctx corectx.Context, warehouseId string, code string,
) (*models.InventoryLocation, error) {
	engine, err := engineFor(models.InventoryLocationSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			models.InventoryLocationFieldWarehouseId, dmodel.Equals, warehouseId),
		*dmodel.NewSearchNode().NewCondition(
			models.InventoryLocationFieldCode, dmodel.Equals, code),
	)

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  1,
	})
	if err != nil {
		return nil, errors.Wrap(err, "FindWarehouseLocationByCode")
	}
	if found == nil || !found.HasData || len(found.Data.Items) == 0 {
		return nil, nil
	}
	return models.NewInventoryLocationFrom(found.Data.Items[0]), nil
}

// WithWarehouseTransaction runs body inside one transaction on a cloned context. Warehouse creation
// and flow reconfiguration each write a warehouse and its locations, and half of either result is
// useless.
func WithWarehouseTransaction(ctx corectx.Context, body func(tranxCtx corectx.Context) error) error {
	engine, err := engineFor(models.WarehouseSchemaName)
	if err != nil {
		return err
	}

	tranx, err := engine.ResourceRepository().BeginTransaction(ctx)
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
	engine, err := engineFor(models.InventoryLocationSchemaName)
	if err != nil {
		return err
	}

	update := dmodel.DynamicFields{models.InventoryLocationFieldId: locationId}
	for key, value := range fields {
		update[key] = value
	}
	_, err = engine.ResourceRepository().Update(ctx, update)
	return errors.Wrap(err, "writeLocationFieldsDirect")
}
