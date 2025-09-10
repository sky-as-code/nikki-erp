package grant_request

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	"github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type GrantRequestRepository interface {
	Create(ctx crud.Context, grantRequest domain.GrantRequest) (*domain.GrantRequest, error)
	FindById(ctx crud.Context, param FindByIdParam) (*domain.GrantRequest, error)
	FindPendingByReceiverAndTarget(ctx crud.Context, receiverId model.Id, targetId model.Id, targetType domain.GrantRequestTargetType) ([]*domain.GrantRequest, error)
	Update(ctx crud.Context, grantRequest domain.GrantRequest) (*domain.GrantRequest, error)
	Delete(ctx crud.Context, id model.Id) error

	BeginTransaction(ctx crud.Context) (*ent.Tx, error)
}

type FindByIdParam = GetGrantRequestQuery
