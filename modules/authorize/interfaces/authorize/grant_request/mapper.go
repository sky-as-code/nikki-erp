package grant_request

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

func (this *CreateGrantRequestCommand) ToGrantRequest() *domain.GrantRequest {
	grantRequest := &domain.GrantRequest{}
	model.MustCopy(this, grantRequest)

	return grantRequest
}

func (this *CancelGrantRequestCommand) ToGrantRequest() *domain.GrantRequest {
	grantRequest := &domain.GrantRequest{}
	model.MustCopy(this, grantRequest)

	return grantRequest
}

func (this DeleteGrantRequestCommand) ToDomainModel() *domain.GrantRequest {
	grantRequest := &domain.GrantRequest{}
	grantRequest.Id = &this.Id
	return grantRequest
}

func (this *RespondToGrantRequestCommand) ToGrantRequest() *domain.GrantRequest {
	grantRequest := &domain.GrantRequest{}
	model.MustCopy(this, grantRequest)
	
	return grantRequest
}
