package commchannel

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type CommChannelRepository interface {
	Create(ctx crud.Context, commChannel *domain.CommChannel) (*domain.CommChannel, error)
	DeleteHard(ctx crud.Context, param DeleteParam) (int, error)
	FindById(ctx crud.Context, param FindByIdParam) (*domain.CommChannel, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[domain.CommChannel], error)
	Update(ctx crud.Context, commChannel *domain.CommChannel, prevEtag model.Etag) (*domain.CommChannel, error)
}

type DeleteParam = DeleteCommChannelCommand
type FindByIdParam = GetCommChannelByIdQuery
type FindByPartyParam = GetCommChannelsByPartyQuery
type SearchParam struct {
	Predicate *orm.Predicate
	Order     []orm.OrderOption
	Page      int
	Size      int
}
