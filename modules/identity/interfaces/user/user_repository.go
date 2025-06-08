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
	Delete(ctx context.Context, param DeleteUserCommand) error
	FindById(ctx context.Context, param FindByIdParam) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx context.Context, predicate *orm.Predicate, order []orm.OrderOption, opts crud.PagingOptions) (*crud.PagedResult[domain.User], error)
	Update(ctx context.Context, user domain.User) (*domain.User, error)
}

type DeleteUserParam = DeleteUserCommand
type FindByIdParam = GetUserByIdQuery

type OrganizationRepository interface {
	Create(ctx context.Context, organization domain.Organization) (*domain.Organization, error)
	Update(ctx context.Context, organization domain.Organization) (*domain.Organization, error)
	Delete(ctx context.Context, id model.Id) error
	FindById(ctx context.Context, id model.Id) (*domain.Organization, error)
	FindBySlug(ctx context.Context, slug string) (*domain.Organization, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx context.Context, predicate *orm.Predicate, order []orm.OrderOption, opts crud.PagingOptions) (*crud.PagedResult[domain.Organization], error)
}
