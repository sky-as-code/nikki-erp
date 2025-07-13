package organization

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/crud"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

type OrganizationRepository interface {
	Create(ctx context.Context, organization domain.Organization) (*domain.Organization, error)
	DeleteHard(ctx context.Context, id model.Id) error
	DeleteSoft(ctx context.Context, id model.Id) (*domain.Organization, error)
	FindById(ctx context.Context, id model.Id) (*domain.Organization, error)
	FindBySlug(ctx context.Context, query GetOrganizationBySlugQuery) (*domain.Organization, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx context.Context, param SearchParam) (*crud.PagedResult[domain.Organization], error)
	Update(ctx context.Context, organization domain.Organization, prevEtag model.Etag) (*domain.Organization, error)
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
