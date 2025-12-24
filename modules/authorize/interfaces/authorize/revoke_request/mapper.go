package revoke_request

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

func (this CreateRevokeRequestCommand) ToDomainModel() *domain.RevokeRequest {
	revokeRequest := &domain.RevokeRequest{}
	model.MustCopy(this, revokeRequest)

	return revokeRequest
}

func (this CreateBulkRevokeRequestsCommand) ToDomainModels() []*domain.RevokeRequest {
	if this.Items == nil {
		return nil
	}

	return array.Map(this.Items, func(cmd CreateRevokeRequestCommand) *domain.RevokeRequest {
		return cmd.ToDomainModel()
	})
}

func (this DeleteRevokeRequestCommand) ToDomainModel() *domain.RevokeRequest {
	revokeRequest := &domain.RevokeRequest{}
	revokeRequest.Id = &this.Id
	return revokeRequest
}
