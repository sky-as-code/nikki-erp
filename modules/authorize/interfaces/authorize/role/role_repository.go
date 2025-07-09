package role

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

type RoleRepository interface {
	Create(ctx context.Context, role domain.Role) (*domain.Role, error)
	FindByName(ctx context.Context, param FindByNameParam) (*domain.Role, error)
	// FindById(ctx context.Context, param FindByIdParam) (*domain.Resource, error)
	// Update(ctx context.Context, resource domain.Resource) (*domain.Resource, error)
	// ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	// Search(ctx context.Context, param SearchParam) (*crud.PagedResult[domain.Resource], error)
}

// type FindByIdParam = GetResourceByIdQuery
type FindByNameParam = GetRoleByNameCommand

// type SearchParam struct {
// 	Predicate   *orm.Predicate
// 	Order       []orm.OrderOption
// 	Page        int
// 	Size        int
// 	WithActions bool
// }
