package purchaseorder

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain"
)

type PurchaseOrderRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys domain.PurchaseOrder) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []domain.PurchaseOrder) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, input domain.PurchaseOrder) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.PurchaseOrder], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.PurchaseOrder]], error)
	Update(ctx corectx.Context, input domain.PurchaseOrder) (*dyn.OpResult[dyn.MutateResultData], error)
}
