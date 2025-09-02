package role

import (
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type RoleRepository interface {
	Create(ctx crud.Context, role domain.Role) (*domain.Role, error)
	CreateWithEntitlements(ctx crud.Context, role domain.Role, entitlementIds []model.Id) (*domain.Role, error)
	UpdateTx(ctx crud.Context, role domain.Role, prevEtag model.Etag, addEntitlementIds, removeEntitlementIds []model.Id) (*domain.Role, error)
	DeleteHardTx(ctx crud.Context, param DeleteRoleHardParam) (int, error)
	FindByName(ctx crud.Context, param FindByNameParam) (*domain.Role, error)
	FindById(ctx crud.Context, param FindByIdParam) (*domain.Role, error)
	Exist(ctx crud.Context, param ExistRoleParam) (bool, error)
	FindAllBySubject(ctx crud.Context, param FindAllBySubjectParam) ([]domain.Role, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, fault.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[domain.Role], error)
}

type FindByIdParam = GetRoleByIdQuery
type ExistRoleParam = GetRoleByIdQuery
type FindByNameParam = GetRoleByNameCommand
type FindAllBySubjectParam = GetRolesBySubjectQuery
type DeleteRoleHardParam struct {
	Id   model.Id `json:"id"`
	Name string   `json:"name"`
}

type SearchParam struct {
	Predicate        *orm.Predicate
	Order            []orm.OrderOption
	Page             int
	Size             int
	WithEntitlements bool
}
