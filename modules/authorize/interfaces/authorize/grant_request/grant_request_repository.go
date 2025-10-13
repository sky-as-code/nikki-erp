package grant_request

import (
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	"github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type GrantRequestRepository interface {
	Create(ctx crud.Context, grantRequest *domain.GrantRequest) (*domain.GrantRequest, error)
	FindAllByTarget(ctx crud.Context, param FindAllByTargetParam) ([]domain.GrantRequest, error)
	FindById(ctx crud.Context, param FindByIdParam) (*domain.GrantRequest, error)
	FindPendingByReceiverAndTarget(ctx crud.Context, receiverId model.Id, targetId model.Id, targetType domain.GrantRequestTargetType) ([]domain.GrantRequest, error)
	Update(ctx crud.Context, grantRequest *domain.GrantRequest, prevEtag model.Etag) (*domain.GrantRequest, error)
	ConfigTargetFields(ctx crud.Context, grantRequest *domain.GrantRequest, name string, prevEtag model.Etag) error
	Delete(ctx crud.Context, param DeleteParam) (int, error)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[domain.GrantRequest], error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, fault.ValidationErrors)

	BeginTransaction(ctx crud.Context) (*ent.Tx, error)
}

type FindAllByTargetParam = GetGrantRequestsByTargetQuery

type FindByIdParam = GetGrantRequestByIdQuery
type DeleteParam = DeleteGrantRequestCommand

type SearchParam struct {
	Predicate *orm.Predicate
	Order     []orm.OrderOption
	Page      int
	Size      int
}
