package entitlement

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/crud"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

type EntitlementRepository interface {
	Create(ctx context.Context, entitlement domain.Entitlement) (*domain.Entitlement, error)
	Update(ctx context.Context, entitlement domain.Entitlement, prevEtag model.Etag) (*domain.Entitlement, error)
	DeleteHard(ctx context.Context, param DeleteParam) (int, error)
	Exists(ctx context.Context, param FindByIdParam) (bool, error)
	FindById(ctx context.Context, param FindByIdParam) (*domain.Entitlement, error)
	FindByName(ctx context.Context, param FindByNameParam) (*domain.Entitlement, error)
	FindAllByIds(ctx context.Context, param FindAllByIdsParam) ([]*domain.Entitlement, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, fault.ValidationErrors)
	Search(ctx context.Context, param SearchParam) (*crud.PagedResult[*domain.Entitlement], error)
}

type FindByNameParam = GetEntitlementByNameQuery
type FindByIdParam = GetEntitlementByIdQuery
type FindAllByIdsParam = GetAllEntitlementByIdsQuery
type DeleteParam = DeleteEntitlementHardByIdQuery

type SearchParam struct {
	Predicate *orm.Predicate
	Order     []orm.OrderOption
	Page      int
	Size      int
}
