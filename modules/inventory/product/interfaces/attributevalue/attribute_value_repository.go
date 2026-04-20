package attributevalue

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

type AttributeValueRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys domain.AttributeValue) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []domain.AttributeValue) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, attributeValue domain.AttributeValue) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.AttributeValue], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.AttributeValue]], error)
	Update(ctx corectx.Context, attributeValue domain.AttributeValue) (*dyn.OpResult[dyn.MutateResultData], error)
	GetIdsByVariantId(ctx corectx.Context, variantId model.Id) ([]model.Id, error)
}
