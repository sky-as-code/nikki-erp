package entitlement

import (
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type EntitlementRepository interface {
	Create(ctx crud.Context, entitlement domain.Entitlement) (*domain.Entitlement, error)
	Update(ctx crud.Context, entitlement domain.Entitlement, prevEtag model.Etag) (*domain.Entitlement, error)
	DeleteHard(ctx crud.Context, param DeleteParam) (int, error)
	Exists(ctx crud.Context, param FindByIdParam) (bool, error)
	FindById(ctx crud.Context, param FindByIdParam) (*domain.Entitlement, error)
	FindByName(ctx crud.Context, param FindByNameParam) (*domain.Entitlement, error)
	FindAllByIds(ctx crud.Context, param FindAllByIdsParam) ([]domain.Entitlement, error)
	FindByActionExpr(ctx crud.Context, param FindByActionExprParam) (*domain.Entitlement, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, fault.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[domain.Entitlement], error)
}

type FindByNameParam = GetEntitlementByNameQuery
type FindByIdParam = GetEntitlementByIdQuery
type FindAllByIdsParam = GetAllEntitlementByIdsQuery
type FindByActionExprParam = GetEntitlementByActionExprQuery
type DeleteParam = DeleteEntitlementHardByIdQuery

type SearchParam struct {
	Predicate *orm.Predicate
	Order     []orm.OrderOption
	Page      int
	Size      int
}
