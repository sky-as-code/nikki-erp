package services

import (
	"fmt"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// supplyGraphScanLimit bounds the walk through the resupply graph, guarding against a cycle that
// somehow got persisted rather than expressing a real limit on how long a supply chain may be.
const supplyGraphScanLimit = 50

// NewSupplyRelationDomainService derives the supply relation service from the engine's default.
func NewSupplyRelationDomainService(base drif.DynamicResourceService) *SupplyRelationDomainServiceImpl {
	return &SupplyRelationDomainServiceImpl{DynamicResourceService: base}
}

// SupplyRelationDomainServiceImpl keeps the resupply topology sane. A relation only declares who
// may restock whom — it reserves nothing and starts no transfer — so what is guarded is the shape
// of the graph: no self-supply, no duplicates, one default per destination, no cycles.
type SupplyRelationDomainServiceImpl struct {
	drif.DynamicResourceService
}

var _ drif.DynamicResourceService = (*SupplyRelationDomainServiceImpl)(nil)

func (this *SupplyRelationDomainServiceImpl) Create(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	vErrs, err := this.assertRelationValid(ctx, params, "")
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: *vErrs}, nil
	}
	return this.DynamicResourceService.Create(ctx, params)
}

// Update re-checks the same rules, since priority and the default flag can change. Source and
// destination are deliberately not updatable: repointing a relation is really a different relation,
// and rewriting one in place would make the audit trail read as though the old route never existed.
func (this *SupplyRelationDomainServiceImpl) Update(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	relationId := readStringParam(params, models.WarehouseSupplyRelationFieldId)

	prepared := copyFields(params)
	delete(prepared, models.WarehouseSupplyRelationFieldSourceWarehouseId)
	delete(prepared, models.WarehouseSupplyRelationFieldDestinationWarehouseId)

	current, err := loadSupplyRelation(ctx, relationId)
	if err != nil {
		return nil, err
	}
	if current == nil {
		vErrs := ft.NewClientErrors()
		appendSupplyViolation(vErrs, "warehouse_supply_relation.not_found",
			"no supply relation with id '"+relationId+"'")
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	if readBoolParam(prepared, models.WarehouseSupplyRelationFieldIsDefault) {
		destinationId := derefId(current.GetDestinationWarehouseId())
		conflict, err := this.findConflictingDefault(ctx, destinationId, relationId)
		if err != nil {
			return nil, err
		}
		if conflict {
			vErrs := ft.NewClientErrors()
			appendSupplyViolation(vErrs, "warehouse_supply_relation.default_exists",
				"another relation is already the default source for this warehouse")
			return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
		}
	}
	return this.DynamicResourceService.Update(ctx, prepared)
}

// assertRelationValid applies every rule about a new relation.
func (this *SupplyRelationDomainServiceImpl) assertRelationValid(
	ctx corectx.Context, params dmodel.DynamicFields, selfId string,
) (*ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()
	sourceId := readStringParam(params, models.WarehouseSupplyRelationFieldSourceWarehouseId)
	destinationId := readStringParam(params, models.WarehouseSupplyRelationFieldDestinationWarehouseId)

	if sourceId == "" || destinationId == "" {
		appendSupplyViolation(vErrs, "warehouse_supply_relation.endpoints_required",
			"a supply relation needs both a source and a destination warehouse")
		return vErrs, nil
	}
	if sourceId == destinationId {
		appendSupplyViolation(vErrs, "warehouse_supply_relation.same_warehouse",
			"a warehouse cannot supply itself")
		return vErrs, nil
	}

	if vErrs, err := this.assertEndpointsUsable(ctx, sourceId, destinationId, vErrs); err != nil {
		return vErrs, err
	} else if vErrs.Count() > 0 {
		return vErrs, nil
	}

	duplicate, err := this.findDuplicate(ctx, sourceId, destinationId, selfId)
	if err != nil {
		return vErrs, err
	}
	if duplicate {
		appendSupplyViolation(vErrs, "warehouse_supply_relation.duplicate",
			"this supply route already exists")
		return vErrs, nil
	}

	if readBoolParam(params, models.WarehouseSupplyRelationFieldIsDefault) {
		conflict, err := this.findConflictingDefault(ctx, destinationId, selfId)
		if err != nil {
			return vErrs, err
		}
		if conflict {
			appendSupplyViolation(vErrs, "warehouse_supply_relation.default_exists",
				"another relation is already the default source for this warehouse")
			return vErrs, nil
		}
	}

	// A cycle would let replenishment planning chase its own tail: A restocks B, B restocks C, C
	// restocks A, with no warehouse actually holding the goods.
	cyclic, err := this.wouldCreateCycle(ctx, sourceId, destinationId)
	if err != nil {
		return vErrs, err
	}
	if cyclic {
		appendSupplyViolation(vErrs, "warehouse_supply_relation.cycle",
			"this route would make the resupply topology circular")
	}
	return vErrs, nil
}

// assertEndpointsUsable checks both warehouses exist, are unarchived, and share an org.
func (this *SupplyRelationDomainServiceImpl) assertEndpointsUsable(
	ctx corectx.Context, sourceId string, destinationId string, vErrs *ft.ClientErrors,
) (*ft.ClientErrors, error) {
	source, sourceErrs, err := loadWarehouseById(ctx, sourceId)
	if err != nil {
		return vErrs, err
	}
	if sourceErrs.Count() > 0 {
		appendSupplyViolation(vErrs, "warehouse_supply_relation.source_not_found",
			"no warehouse with id '"+sourceId+"'")
		return vErrs, nil
	}
	destination, destErrs, err := loadWarehouseById(ctx, destinationId)
	if err != nil {
		return vErrs, err
	}
	if destErrs.Count() > 0 {
		appendSupplyViolation(vErrs, "warehouse_supply_relation.destination_not_found",
			"no warehouse with id '"+destinationId+"'")
		return vErrs, nil
	}

	if isArchivedWarehouse(*source) || isArchivedWarehouse(*destination) {
		appendSupplyViolation(vErrs, "warehouse_supply_relation.warehouse_archived",
			"an archived warehouse cannot take part in a supply route")
	}
	if derefId(source.GetOrgId()) != derefId(destination.GetOrgId()) {
		appendSupplyViolation(vErrs, "warehouse_supply_relation.org_mismatch",
			"both warehouses must belong to the same organisation")
	}
	return vErrs, nil
}

// wouldCreateCycle walks the existing routes from the destination to see whether the source is
// already downstream of it.
func (this *SupplyRelationDomainServiceImpl) wouldCreateCycle(
	ctx corectx.Context, sourceId string, destinationId string,
) (bool, error) {
	visited := map[string]bool{}
	frontier := []string{destinationId}

	for hops := 0; len(frontier) > 0; hops++ {
		if hops >= supplyGraphScanLimit {
			return false, errors.New("the resupply topology is deeper than " +
				fmt.Sprint(supplyGraphScanLimit) + " levels, or already contains a cycle")
		}

		next := make([]string, 0)
		for _, warehouseId := range frontier {
			if visited[warehouseId] {
				continue
			}
			visited[warehouseId] = true
			if warehouseId == sourceId {
				return true, nil
			}

			supplied, err := this.listSuppliedWarehouses(ctx, warehouseId)
			if err != nil {
				return false, err
			}
			next = append(next, supplied...)
		}
		frontier = next
	}
	return false, nil
}

func isArchivedWarehouse(warehouse models.Warehouse) bool {
	archived := warehouse.GetIsArchived()
	return archived != nil && *archived
}

func appendSupplyViolation(vErrs *ft.ClientErrors, key string, message string) {
	vErrs.Append(*ft.NewBusinessViolation(models.WarehouseSupplyRelationSchemaName, key, message))
}
