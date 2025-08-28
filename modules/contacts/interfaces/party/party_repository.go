package party

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/crud"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
)

type PartyRepository interface {
	Create(ctx context.Context, party domain.Party) (*domain.Party, error)
	DeleteHard(ctx context.Context, param DeleteParam) (int, error)
	FindById(ctx context.Context, param FindByIdParam) (*domain.Party, error)
	FindByDisplayName(ctx context.Context, param FindByDisplayNameParam) (*domain.Party, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx context.Context, param SearchParam) (*crud.PagedResult[domain.Party], error)
	Update(ctx context.Context, party domain.Party, prevEtag model.Etag) (*domain.Party, error)
}

type DeleteParam = DeletePartyCommand
type FindByIdParam = GetPartyByIdQuery
type FindByDisplayNameParam = GetPartyByDisplayNameQuery
type SearchParam struct {
	Predicate         *orm.Predicate
	Order             []orm.OrderOption
	Page              int
	Size              int
	WithCommChannels  bool
	WithRelationships bool
}
