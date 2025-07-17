package role_suite

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/crud"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

type RoleSuiteRepository interface {
	Create(ctx context.Context, roleSuite domain.RoleSuite) (*domain.RoleSuite, error)
	FindByName(ctx context.Context, param FindByNameParam) (*domain.RoleSuite, error)
	FindById(ctx context.Context, param FindByIdParam) (*domain.RoleSuite, error)
	FindAllBySubject(ctx context.Context, param FindAllBySubjectParam) ([]domain.RoleSuite, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx context.Context, param SearchParam) (*crud.PagedResult[domain.RoleSuite], error)
}

type FindByIdParam = GetRoleSuiteByIdQuery
type FindByNameParam = GetRoleSuiteByNameCommand
type FindAllBySubjectParam = GetRoleSuitesBySubjectQuery

type SearchParam struct {
	Predicate *orm.Predicate
	Order     []orm.OrderOption
	Page      int
	Size      int
}
