package unitcategory

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/inventory/unit/domain"
)

type UnitCategoryRepository interface {
	Create(ctx crud.Context, unitCategory *domain.UnitCategory) (*domain.UnitCategory, error)
	Update(ctx crud.Context, unitCategory *domain.UnitCategory, prevEtag model.Etag) (*domain.UnitCategory, error)
	DeleteById(ctx crud.Context, id model.Id) (int, error)
	FindById(ctx crud.Context, query FindByIdParam) (*domain.UnitCategory, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[domain.UnitCategory], error)
}

type UnitCategoryService interface {
	CreateUnitCategory(ctx crud.Context, cmd CreateUnitCategoryCommand) (*CreateUnitCategoryResult, error)
	UpdateUnitCategory(ctx crud.Context, cmd UpdateUnitCategoryCommand) (*UpdateUnitCategoryResult, error)
	DeleteUnitCategory(ctx crud.Context, cmd DeleteUnitCategoryCommand) (*DeleteUnitCategoryResult, error)
	GetUnitCategoryById(ctx crud.Context, query GetUnitCategoryByIdQuery) (*GetUnitCategoryByIdResult, error)
	SearchUnitCategories(ctx crud.Context, query SearchUnitCategoriesQuery) (*SearchUnitCategoriesResult, error)
}

type DeleteParam = DeleteUnitCategoryCommand
type FindByIdParam = GetUnitCategoryByIdQuery
type SearchParam struct {
	Predicate *orm.Predicate
	Order     []orm.OrderOption
	Page      int
	Size      int
}
