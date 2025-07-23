package role

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/crud"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

type RoleRepository interface {
	Create(ctx context.Context, role domain.Role) (*domain.Role, error)
	CreateWithEntitlements(ctx context.Context, role domain.Role, entitlementIds []model.Id) (*domain.Role, error)
	FindByName(ctx context.Context, param FindByNameParam) (*domain.Role, error)
	FindById(ctx context.Context, param FindByIdParam) (*domain.Role, error)
	FindAllBySubject(ctx context.Context, param FindAllBySubjectParam) ([]*domain.Role, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, fault.ValidationErrors)
	Search(ctx context.Context, param SearchParam) (*crud.PagedResult[*domain.Role], error)
}

type FindByIdParam = GetRoleByIdQuery
type FindByNameParam = GetRoleByNameCommand
type FindAllBySubjectParam = GetRolesBySubjectQuery

type SearchParam struct {
	Predicate        *orm.Predicate
	Order            []orm.OrderOption
	Page             int
	Size             int
	WithEntitlements bool
}
