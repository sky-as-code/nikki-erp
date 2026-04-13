package productcategory

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

type ProductCategoryRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys domain.ProductCategory) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []domain.ProductCategory) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, productCategory domain.ProductCategory) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.ProductCategory], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.ProductCategory]], error)
	Update(ctx corectx.Context, productCategory domain.ProductCategory) (*dyn.OpResult[dyn.MutateResultData], error)
}
