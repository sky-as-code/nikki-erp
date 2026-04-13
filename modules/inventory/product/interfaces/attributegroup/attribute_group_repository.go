package attributegroup

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

type AttributeGroupRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys domain.AttributeGroup) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []domain.AttributeGroup) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, attributeGroup domain.AttributeGroup) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.AttributeGroup], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.AttributeGroup]], error)
	Update(ctx corectx.Context, attributeGroup domain.AttributeGroup) (*dyn.OpResult[dyn.MutateResultData], error)
}
