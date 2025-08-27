package role_suite

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type RoleSuiteRepository interface {
	Create(ctx context.Context, roleSuite domain.RoleSuite, roleIds []model.Id) (*domain.RoleSuite, error)
	UpdateTx(ctx context.Context, roleSuite domain.RoleSuite, prevEtag model.Etag, addRoleIds, removeRoleIds []model.Id) (*domain.RoleSuite, error)
	DeleteHardTx(ctx context.Context, param DeleteRoleSuiteParam) (int, error)
	FindByName(ctx context.Context, param FindByNameParam) (*domain.RoleSuite, error)
	FindById(ctx context.Context, param FindByIdParam) (*domain.RoleSuite, error)
	FindAllBySubject(ctx context.Context, param FindAllBySubjectParam) ([]domain.RoleSuite, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, fault.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[domain.RoleSuite], error)
}

type FindByIdParam = GetRoleSuiteByIdQuery
type FindByNameParam = GetRoleSuiteByNameCommand
type FindAllBySubjectParam = GetRoleSuitesBySubjectQuery

type DeleteRoleSuiteParam struct {
	Id   model.Id `json:"id"`
	Name string   `json:"name"`
}

type SearchParam struct {
	Predicate *orm.Predicate
	Order     []orm.OrderOption
	Page      int
	Size      int
}
