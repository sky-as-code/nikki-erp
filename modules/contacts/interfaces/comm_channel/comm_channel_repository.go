package comm_channel

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/crud"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
)

type CommChannelRepository interface {
	Create(ctx context.Context, commChannel domain.CommChannel) (*domain.CommChannel, error)
	DeleteHard(ctx context.Context, param DeleteParam) (int, error)
	FindById(ctx context.Context, param FindByIdParam) (*domain.CommChannel, error)
	FindByParty(ctx context.Context, param FindByPartyParam) ([]*domain.CommChannel, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx context.Context, param SearchParam) (*crud.PagedResult[domain.CommChannel], error)
	Update(ctx context.Context, commChannel domain.CommChannel, prevEtag model.Etag) (*domain.CommChannel, error)
}

type DeleteParam = DeleteCommChannelCommand
type FindByIdParam = GetCommChannelByIdQuery
type FindByPartyParam = GetCommChannelsByPartyQuery
type SearchParam struct {
	Predicate *orm.Predicate
	Order     []orm.OrderOption
	Page      int
	Size      int
	WithParty bool
}
