package action

import (
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type ActionRepository interface {
	Create(ctx crud.Context, action domain.Action) (*domain.Action, error)
	FindById(ctx crud.Context, param FindByIdParam) (*domain.Action, error)
	FindByName(ctx crud.Context, param FindByNameParam) (*domain.Action, error)
	Update(ctx crud.Context, action domain.Action, prevEtag model.Etag) (*domain.Action, error)
	DeleteHard(ctx crud.Context, param DeleteParam) (int, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, fault.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[domain.Action], error)
}

type FindByIdParam = GetActionByIdQuery
type FindByNameParam = GetActionByNameCommand
type DeleteParam = DeleteActionHardByIdQuery

type SearchParam struct {
	Predicate *orm.Predicate
	Order     []orm.OrderOption
	Page      int
	Size      int
}
