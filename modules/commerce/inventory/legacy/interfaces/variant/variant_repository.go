package variant

import (
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/legacy/domain"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type VariantRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys domain.Variant) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []domain.Variant) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, variant domain.Variant) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.Variant], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.Variant]], error)
	Update(ctx corectx.Context, variant domain.Variant) (*dyn.OpResult[dyn.MutateResultData], error)
}
