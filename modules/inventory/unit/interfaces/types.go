package interfaces

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type UnitRepository interface {
	Create(ctx crud.Context, unit *Unit) (*Unit, error)
	Update(ctx crud.Context, unit *Unit, prevEtag model.Etag) (*Unit, error)
	DeleteById(ctx crud.Context, id model.Id) (int, error)
	FindById(ctx crud.Context, query FindByIdParam) (*Unit, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[Unit], error)
}

type UnitService interface {
	CreateUnit(ctx crud.Context, cmd CreateUnitCommand) (*CreateUnitResult, error)
	UpdateUnit(ctx crud.Context, cmd UpdateUnitCommand) (*UpdateUnitResult, error)
	DeleteUnit(ctx crud.Context, cmd DeleteUnitCommand) (*DeleteUnitResult, error)
	GetUnitById(ctx crud.Context, query GetUnitByIdQuery) (*GetUnitByIdResult, error)
	SearchUnits(ctx crud.Context, query SearchUnitsQuery) (*SearchUnitsResult, error)
}

type DeleteParam = DeleteUnitCommand
type FindByIdParam = GetUnitByIdQuery
type SearchParam struct {
	Predicate *orm.Predicate
	Order     []orm.OrderOption
	Page      int
	Size      int
}
