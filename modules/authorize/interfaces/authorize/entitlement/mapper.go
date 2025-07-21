package entitlement

import (
	"github.com/sky-as-code/nikki-erp/common/model"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

func (this CreateEntitlementCommand) ToEntitlement() *domain.Entitlement {
	return &domain.Entitlement{
		Name:        &this.Name,
		Description: this.Description,
		ActionId:    this.ActionId,
		ResourceId:  this.ResourceId,
		ScopeRef:    this.ScopeRef,
		ActionExpr:  &this.ActionExpr,
		CreatedBy:   &this.CreatedBy,
	}
}

func (this UpdateEntitlementCommand) ToEntitlement() *domain.Entitlement {
	return &domain.Entitlement{
		ModelBase: model.ModelBase{
			Id:   &this.Id,
			Etag: &this.Etag,
		},
		Description: this.Description,
	}
}
