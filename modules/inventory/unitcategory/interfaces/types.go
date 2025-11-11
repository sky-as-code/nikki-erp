package interfaces

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type UnitCategoryRepository interface {
	Create(ctx crud.Context, unitCategory *UnitCategory) (*UnitCategory, error)
	Update(ctx crud.Context, unitCategory *UnitCategory, prevEtag model.Etag) (*UnitCategory, error)
	DeleteById(ctx crud.Context, id model.Id) (int, error)
	FindById(ctx crud.Context, query FindByIdParam) (*UnitCategory, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[UnitCategory], error)
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
