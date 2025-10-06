package organization

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

type OrganizationRepository interface {
	Create(ctx crud.Context, organization *domain.Organization) (*domain.Organization, error)
	DeleteHard(ctx crud.Context, param DeleteParam) (int, error)
	DeleteSoft(ctx crud.Context, param DeleteParam) (*domain.Organization, error)
	FindById(ctx crud.Context, id model.Id) (*domain.Organization, error)
	FindBySlug(ctx crud.Context, query FindBySlugParam) (*domain.Organization, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[domain.Organization], error)
	Update(ctx crud.Context, organization domain.Organization, prevEtag model.Etag) (*domain.Organization, error)
	Exists(ctx crud.Context, id model.Id) (bool, error)
}

type DeleteParam = DeleteOrganizationCommand
type FindBySlugParam = GetOrganizationBySlugQuery
type SearchParam struct {
	Predicate      *orm.Predicate
	Order          []orm.OrderOption
	Page           int
	Size           int
	IncludeDeleted bool
}
