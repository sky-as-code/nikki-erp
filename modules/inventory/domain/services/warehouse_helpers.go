package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// loadWarehouse fetches one warehouse, reporting a missing id as a client error: the id came from
// the caller.
func (this *WarehouseDomainServiceImpl) loadWarehouse(
	ctx corectx.Context, warehouseId string,
) (*models.Warehouse, *ft.ClientErrors, error) {
	return loadWarehouseById(ctx, warehouseId)
}

// loadWarehouseById is the package-level read, so the application service can resolve a warehouse
// without holding the domain service.
func loadWarehouseById(
	ctx corectx.Context, warehouseId string,
) (*models.Warehouse, *ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()
	if warehouseId == "" {
		appendWarehouseViolation(vErrs, "warehouse.not_found", "no warehouse id was given")
		return nil, vErrs, nil
	}

	engine, err := engineFor(models.WarehouseSchemaName)
	if err != nil {
		return nil, vErrs, err
	}
	found, err := engine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{
		models.WarehouseFieldId: warehouseId,
	})
	if err != nil {
		return nil, vErrs, errors.Wrap(err, "loadWarehouseById")
	}
	if found == nil || !found.HasData {
		appendWarehouseViolation(vErrs, "warehouse.not_found",
			"no warehouse with id '"+warehouseId+"'")
		return nil, vErrs, nil
	}
	return models.NewWarehouseFrom(found.Data), vErrs, nil
}

// WriteStatus sets a warehouse's operational state.
//
// It goes through the repository rather than the service, so the rules the caller already checked
// are not re-run and the write cannot recurse into the overridden Update.
func (this *WarehouseDomainServiceImpl) WriteStatus(
	ctx corectx.Context, warehouseId string, status string,
) error {
	return writeWarehouseFields(ctx, warehouseId, dmodel.DynamicFields{
		models.WarehouseFieldStatus: status,
	})
}

func writeWarehouseFields(
	ctx corectx.Context, warehouseId string, fields dmodel.DynamicFields,
) error {
	engine, err := engineFor(models.WarehouseSchemaName)
	if err != nil {
		return err
	}

	update := dmodel.DynamicFields{models.WarehouseFieldId: warehouseId}
	for key, value := range fields {
		update[key] = value
	}
	_, err = engine.ResourceRepository().Update(ctx, update)
	return errors.Wrap(err, "writeWarehouseFields")
}

// countUnarchivedChildren reports how many live warehouses sit directly under one.
func (this *WarehouseDomainServiceImpl) countUnarchivedChildren(
	ctx corectx.Context, warehouseId string,
) (int, error) {
	engine, err := engineFor(models.WarehouseSchemaName)
	if err != nil {
		return 0, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			models.WarehouseFieldParentWarehouseId, dmodel.Equals, warehouseId),
	)

	total, err := countMatching(ctx, engine, graph)
	return total, errors.Wrap(err, "countUnarchivedChildren")
}

// countUnarchivedSupplyRelations reports how many live resupply routes name this warehouse, in
// either direction.
func (this *WarehouseDomainServiceImpl) countUnarchivedSupplyRelations(
	ctx corectx.Context, warehouseId string,
) (int, error) {
	engine, err := engineFor(models.WarehouseSupplyRelationSchemaName)
	if err != nil {
		return 0, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().Or(
			*dmodel.NewSearchNode().NewCondition(
				models.WarehouseSupplyRelationFieldSourceWarehouseId, dmodel.Equals, warehouseId),
			*dmodel.NewSearchNode().NewCondition(
				models.WarehouseSupplyRelationFieldDestinationWarehouseId, dmodel.Equals, warehouseId),
		),
	)

	total, err := countMatching(ctx, engine, graph)
	return total, errors.Wrap(err, "countUnarchivedSupplyRelations")
}

func appendWarehouseViolation(vErrs *ft.ClientErrors, key string, message string) {
	vErrs.Append(*ft.NewBusinessViolation(models.WarehouseSchemaName, key, message))
}

func warehouseViolationResult(key string, message string) *dyn.OpResult[dyn.MutateResultData] {
	vErrs := ft.NewClientErrors()
	appendWarehouseViolation(vErrs, key, message)
	return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}
}
