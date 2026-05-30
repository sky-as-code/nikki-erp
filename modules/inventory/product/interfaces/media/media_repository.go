package media

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

type InventoryMediaRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys domain.InventoryMedia) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []domain.InventoryMedia) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, media domain.InventoryMedia) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.InventoryMedia], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.InventoryMedia]], error)
	Update(ctx corectx.Context, media domain.InventoryMedia) (*dyn.OpResult[dyn.MutateResultData], error)
}
