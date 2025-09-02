package party

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type PartyRepository interface {
	Create(ctx crud.Context, party domain.Party) (*domain.Party, error)
	DeleteHard(ctx crud.Context, param DeleteParam) (int, error)
	FindById(ctx crud.Context, param FindByIdParam) (*domain.Party, error)
	FindByDisplayName(ctx crud.Context, param FindByDisplayNameParam) (*domain.Party, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[domain.Party], error)
	Update(ctx crud.Context, party domain.Party, prevEtag model.Etag) (*domain.Party, error)
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
