package revoke_request

import (
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	"github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type RevokeRequestRepository interface {
	Create(ctx crud.Context, revokeRequest *domain.RevokeRequest) (*domain.RevokeRequest, error)

	BeginTransaction(ctx crud.Context) (*ent.Tx, error)
}
