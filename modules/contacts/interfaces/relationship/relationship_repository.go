package relationship

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type RelationshipRepository interface {
	Create(ctx crud.Context, relationship domain.Relationship) (*domain.Relationship, error)
	DeleteHard(ctx crud.Context, param DeleteParam) (int, error)
	FindById(ctx crud.Context, param FindByIdParam) (*domain.Relationship, error)
	FindByParty(ctx crud.Context, param FindByPartyParam) ([]*domain.Relationship, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[domain.Relationship], error)
	Update(ctx crud.Context, relationship domain.Relationship, prevEtag model.Etag) (*domain.Relationship, error)
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
