package user

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/crud"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) (*domain.User, error)
	DeleteHard(ctx context.Context, param DeleteParam) (int, error)
	Exists(ctx context.Context, id model.Id) (bool, error)
	ExistsMulti(ctx context.Context, ids []model.Id) (existing []model.Id, notExisting []model.Id, err error)
	FindById(ctx context.Context, param FindByIdParam) (*domain.User, error)
	FindByEmail(ctx context.Context, param FindByEmailParam) (*domain.User, error)
	FindUsersByHierarchyId(ctx context.Context, param FindByHierarchyIdParam) ([]domain.User, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx context.Context, param SearchParam) (*crud.PagedResult[domain.User], error)
	Update(ctx context.Context, user domain.User, prevEtag model.Etag) (*domain.User, error)
}

type DeleteParam = DeleteUserCommand
type ExistsParam = UserExistsCommand
type ExistsMultiParam = UserExistsMultiCommand
type FindByIdParam = GetUserByIdQuery
type FindByEmailParam = GetUserByEmailQuery
type FindByHierarchyIdParam struct {
	HierarchyId model.Id
}
type SearchParam struct {
	Predicate  *orm.Predicate
	Order      []orm.OrderOption
	Page       int
	Size       int
	WithGroups bool
}
