package grant_request

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

type GrantRequestRepository interface {
	Create(ctx context.Context, grantRequest domain.GrantRequest) (*domain.GrantRequest, error)
	FindById(ctx context.Context, param FindByIdParam) (*domain.GrantRequest, error)
	FindPendingByReceiverAndTarget(ctx context.Context, receiverId model.Id, targetId model.Id, targetType domain.GrantRequestTargetType) ([]*domain.GrantRequest, error)
	Update(ctx context.Context, grantRequest domain.GrantRequest) (*domain.GrantRequest, error)
	Delete(ctx context.Context, id model.Id) error
}

type FindByIdParam = GetGrantRequestQuery
