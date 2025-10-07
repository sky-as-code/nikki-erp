package revoke_request

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

func (this CreateRevokeRequestCommand) ToDomainModel() *domain.RevokeRequest {
	revokeRequest := &domain.RevokeRequest{}
	model.MustCopy(this, revokeRequest)
	return revokeRequest
}
