package grant_response

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

type GrantResponseRepository interface {
	Create(ctx context.Context, grantResponse domain.GrantResponse) (*domain.GrantResponse, error)
	// FindById(ctx context.Context, id model.Id) (*domain.GrantResponse, error)
	// FindByRequestId(ctx context.Context, requestId model.Id) ([]*domain.GrantResponse, error)
}
