package group

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

type GroupRepository interface {
	AddRemoveUsers(ctx crud.Context, param AddRemoveUsersParam) (*ft.ClientError, error)
	Create(ctx crud.Context, group domain.Group) (*domain.Group, error)
	DeleteHard(ctx crud.Context, param DeleteParam) (int, error)
	FindById(ctx crud.Context, param FindByIdParam) (*domain.Group, error)
	FindByName(ctx crud.Context, param FindByNameParam) (*domain.Group, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[domain.Group], error)
	Update(ctx crud.Context, group domain.Group, prevEtag model.Etag) (*domain.Group, error)
	Exists(ctx crud.Context, param ExistsParam) (bool, error)
}

type AddRemoveUsersParam = AddRemoveUsersCommand
type DeleteParam = DeleteGroupCommand
type FindByIdParam = GetGroupByIdQuery
type ExistsParam = GroupExistsCommand
type FindByNameParam struct {
	Name string
}
type SearchParam struct {
	Predicate *orm.Predicate
	Order     []orm.OrderOption
	Page      int
	Size      int
	WithOrg   bool
}
