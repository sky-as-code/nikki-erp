package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The reads behind the supply relation rules. All ignore archived relations: an archived route
// neither blocks a new one nor extends the graph.

func loadSupplyRelation(
	ctx corectx.Context, relationId string,
) (*models.WarehouseSupplyRelation, error) {
	if relationId == "" {
		return nil, nil
	}

	engine, err := engineFor(models.WarehouseSupplyRelationSchemaName)
	if err != nil {
		return nil, err
	}
	found, err := engine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{
		models.WarehouseSupplyRelationFieldId: relationId,
	})
	if err != nil {
		return nil, errors.Wrap(err, "loadSupplyRelation")
	}
	if found == nil || !found.HasData {
		return nil, nil
	}
	return models.NewWarehouseSupplyRelationFrom(found.Data), nil
}

// findDuplicate reports whether the same route already exists.
func (this *SupplyRelationDomainServiceImpl) findDuplicate(
	ctx corectx.Context, sourceId string, destinationId string, selfId string,
) (bool, error) {
	engine, err := engineFor(models.WarehouseSupplyRelationSchemaName)
	if err != nil {
		return false, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			models.WarehouseSupplyRelationFieldSourceWarehouseId, dmodel.Equals, sourceId),
		*dmodel.NewSearchNode().NewCondition(
			models.WarehouseSupplyRelationFieldDestinationWarehouseId, dmodel.Equals, destinationId),
	)

	matches, err := this.searchRelations(ctx, engine, graph)
	if err != nil {
		return false, err
	}
	for _, match := range matches {
		if derefString(match.GetId()) != selfId {
			return true, nil
		}
	}
	return false, nil
}

// findConflictingDefault reports whether some other live relation is already the default source
// for a destination. A destination may have many sources but only one default.
func (this *SupplyRelationDomainServiceImpl) findConflictingDefault(
	ctx corectx.Context, destinationId string, selfId string,
) (bool, error) {
	engine, err := engineFor(models.WarehouseSupplyRelationSchemaName)
	if err != nil {
		return false, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			models.WarehouseSupplyRelationFieldDestinationWarehouseId, dmodel.Equals, destinationId),
		*dmodel.NewSearchNode().NewCondition(
			models.WarehouseSupplyRelationFieldIsDefault, dmodel.Equals, true),
	)

	matches, err := this.searchRelations(ctx, engine, graph)
	if err != nil {
		return false, err
	}
	for _, match := range matches {
		if derefString(match.GetId()) != selfId {
			return true, nil
		}
	}
	return false, nil
}

// listSuppliedWarehouses returns the warehouses a given one supplies, for the cycle walk.
func (this *SupplyRelationDomainServiceImpl) listSuppliedWarehouses(
	ctx corectx.Context, sourceId string,
) ([]string, error) {
	engine, err := engineFor(models.WarehouseSupplyRelationSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			models.WarehouseSupplyRelationFieldSourceWarehouseId, dmodel.Equals, sourceId),
	)

	matches, err := this.searchRelations(ctx, engine, graph)
	if err != nil {
		return nil, err
	}

	supplied := make([]string, 0, len(matches))
	for _, match := range matches {
		supplied = append(supplied, derefId(match.GetDestinationWarehouseId()))
	}
	return supplied, nil
}

func (this *SupplyRelationDomainServiceImpl) searchRelations(
	ctx corectx.Context, engine drif.DynamicResourceEngine, graph *dmodel.SearchGraph,
) ([]models.WarehouseSupplyRelation, error) {
	relations := make([]models.WarehouseSupplyRelation, 0)
	for page := 0; ; page++ {
		found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
			Graph: graph,
			Page:  page,
			Size:  usageScanPageSize,
		})
		if err != nil {
			return nil, errors.Wrap(err, "searchRelations")
		}
		if found == nil || !found.HasData || len(found.Data.Items) == 0 {
			break
		}
		for _, row := range found.Data.Items {
			relations = append(relations, *models.NewWarehouseSupplyRelationFrom(row))
		}
		if len(found.Data.Items) < usageScanPageSize {
			break
		}
	}
	return relations, nil
}
