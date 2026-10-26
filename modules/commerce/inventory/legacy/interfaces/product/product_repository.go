package product

import (
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/legacy/domain"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type ProductRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys domain.Product) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []domain.Product) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, product domain.Product) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.Product], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.Product]], error)
	Update(ctx corectx.Context, product domain.Product) (*dyn.OpResult[dyn.MutateResultData], error)
}
