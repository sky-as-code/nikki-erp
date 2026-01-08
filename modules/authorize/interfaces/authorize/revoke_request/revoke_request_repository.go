package revoke_request

import (
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	"github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type RevokeRequestRepository interface {
	Create(ctx crud.Context, revokeRequest *domain.RevokeRequest) (*domain.RevokeRequest, error)
	CreateBulk(ctx crud.Context, revokeRequests []*domain.RevokeRequest) ([]*domain.RevokeRequest, error)
	FindById(ctx crud.Context, param FindByIdParam) (*domain.RevokeRequest, error)
	FindAllByTarget(ctx crud.Context, param FindAllByTargetParam) ([]domain.RevokeRequest, error)
	UpdateTargetFields(ctx crud.Context, revokeRequest *domain.RevokeRequest, prevEtag model.Etag) error
	Delete(ctx crud.Context, param DeleteParam) (int, error)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[domain.RevokeRequest], error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, fault.ValidationErrors)

	BeginTransaction(ctx crud.Context) (*ent.Tx, error)
}

type FindByIdParam = GetRevokeRequestByIdQuery
type DeleteParam = DeleteRevokeRequestCommand

type FindAllByTargetParam struct {
	TargetType domain.GrantRequestTargetType
	TargetRef  model.Id
}

type SearchParam struct {
	Predicate *orm.Predicate
	Order     []orm.OrderOption
	Page      int
	Size      int
}
