package entitlement

import (
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

func (this CreateEntitlementCommand) ToEntitlement() *domain.Entitlement {
	return &domain.Entitlement{
		ActionId:    this.ActionId,
		ActionExpr:  &this.ActionExpr,
		Name:        &this.Name,
		Description: this.Description,
		ResourceId:  this.ResourceId,
		// SubjectType: &this.SubjectType,
		// SubjectRef:  &this.SubjectRef,
		ScopeRef:    this.ScopeRef,
		CreatedBy:   &this.CreatedBy,
	}
}

// func (this UpdateResourceCommand) ToResource() *domain.Resource {
// 	return &domain.Resource{
// 		ModelBase: model.ModelBase{
// 			Id:   &this.Id,
// 			Etag: &this.Etag,
// 		},
// 		Description: this.Description,
// 	}
// }
