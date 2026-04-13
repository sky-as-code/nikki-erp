package unitcategory

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/unit/domain"
)

type UnitCategoryRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys domain.UnitCategory) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []domain.UnitCategory) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, unitCategory domain.UnitCategory) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.UnitCategory], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.UnitCategory]], error)
	Update(ctx corectx.Context, unitCategory domain.UnitCategory) (*dyn.OpResult[dyn.MutateResultData], error)
}

type UnitCategoryService interface {
	CreateUnitCategory(ctx corectx.Context, cmd CreateUnitCategoryCommand) (*CreateUnitCategoryResult, error)
	UpdateUnitCategory(ctx corectx.Context, cmd UpdateUnitCategoryCommand) (*UpdateUnitCategoryResult, error)
	DeleteUnitCategory(ctx corectx.Context, cmd DeleteUnitCategoryCommand) (*DeleteUnitCategoryResult, error)
	GetUnitCategory(ctx corectx.Context, query GetUnitCategoryQuery) (*GetUnitCategoryResult, error)
	SearchUnitCategories(ctx corectx.Context, query SearchUnitCategoriesQuery) (*SearchUnitCategoriesResult, error)
	UnitCategoryExists(ctx corectx.Context, query UnitCategoryExistsQuery) (*UnitCategoryExistsResult, error)
}
