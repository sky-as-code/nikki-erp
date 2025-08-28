package relationship

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/crud"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
)

type RelationshipRepository interface {
	Create(ctx context.Context, relationship domain.Relationship) (*domain.Relationship, error)
	DeleteHard(ctx context.Context, param DeleteParam) (int, error)
	FindById(ctx context.Context, param FindByIdParam) (*domain.Relationship, error)
	FindByParty(ctx context.Context, param FindByPartyParam) ([]*domain.Relationship, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx context.Context, param SearchParam) (*crud.PagedResult[domain.Relationship], error)
	Update(ctx context.Context, relationship domain.Relationship, prevEtag model.Etag) (*domain.Relationship, error)
}

type DeleteParam = DeleteRelationshipCommand
type FindByIdParam = GetRelationshipByIdQuery
type FindByPartyParam = GetRelationshipsByPartyQuery
type SearchParam struct {
	Predicate       *orm.Predicate
	Order           []orm.OrderOption
	Page            int
	Size            int
	WithTargetParty bool
}
