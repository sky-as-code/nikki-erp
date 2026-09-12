package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/warehouse"
)

var _ itWarehouse.WarehouseReadService = (*WarehouseDomainServiceImpl)(nil)

// SearchWarehouses is the org-scoped read a caller with no user permission of its own reaches
// through, so the org condition is ANDed above whatever graph the caller sent — narrowing the
// result, never widening it past the org.
func (this *WarehouseDomainServiceImpl) SearchWarehouses(
	ctx corectx.Context, query itWarehouse.SearchWarehousesReadQuery,
) (*itWarehouse.SearchWarehousesReadResult, error) {
	scoped := query.SearchQuery
	scoped.Graph = dmodel.MergeAndCondition(
		scoped.Graph, models.WarehouseFieldOrgId, dmodel.Equals, string(query.OrgId),
	)

	result, err := this.Search(ctx, searchQueryToParams(scoped))
	if err != nil {
		return nil, errors.Wrap(err, "SearchWarehouses")
	}
	if result.ClientErrors.Count() > 0 {
		return &itWarehouse.SearchWarehousesReadResult{ClientErrors: result.ClientErrors}, nil
	}
	if !result.HasData {
		return emptyPage[models.Warehouse](scoped), nil
	}
	return pagedResult(&result.Data, scoped, models.NewWarehouseFrom), nil
}
