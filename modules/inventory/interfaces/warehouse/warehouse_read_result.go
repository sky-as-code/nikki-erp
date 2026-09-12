package warehouse

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// SearchWarehousesReadQuery is the org-scoped counterpart of SearchWarehousesQuery (the composable
// CRUD query), for a caller that reaches this narrower read port with no user permission of its
// own — a device certificate proves which machine it is, and nothing about who is operating it —
// and so must supply the org itself rather than have it asserted from the caller's membership.
type SearchWarehousesReadQuery struct {
	OrgId model.Id

	dyn.SearchQuery
}

type SearchWarehousesResultData = dyn.PagedResultData[models.Warehouse]
type SearchWarehousesReadResult = dyn.OpResult[SearchWarehousesResultData]

// WarehouseReadService lists the warehouses a caller offers for selection. Kept apart from
// WarehouseAppService: listing what exists grants none of the power to create one or reshape its
// flows.
type WarehouseReadService interface {
	SearchWarehouses(ctx corectx.Context, query SearchWarehousesReadQuery) (*SearchWarehousesReadResult, error)
}
