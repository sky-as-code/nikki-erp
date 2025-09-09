package group

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/crud"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

type GroupRepository interface {
	AddRemoveUsers(ctx context.Context, param AddRemoveUsersParam) (*ft.ClientError, error)
	Create(ctx context.Context, group domain.Group) (*domain.Group, error)
	DeleteHard(ctx context.Context, param DeleteParam) (int, error)
	FindById(ctx context.Context, param FindByIdParam) (*domain.Group, error)
	FindByName(ctx context.Context, param FindByNameParam) (*domain.Group, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx context.Context, param SearchParam) (*crud.PagedResult[domain.Group], error)
	Update(ctx context.Context, group domain.Group, prevEtag model.Etag) (*domain.Group, error)
	Exists(ctx context.Context, param ExistsParam) (bool, error)
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
