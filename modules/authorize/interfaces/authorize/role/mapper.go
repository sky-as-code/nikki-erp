package role

import (
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

func (this CreateRoleCommand) ToRole() *domain.Role {
	return &domain.Role{
		Name:                 &this.Name,
		Description:          this.Description,
		OwnerType:            domain.WrapRoleOwnerType(this.OwnerType),
		OwnerRef:             &this.OwnerRef,
		IsRequestable:        &this.IsRequestable,
		IsRequiredAttachment: &this.IsRequiredAttachment,
		IsRequiredComment:    &this.IsRequiredComment,
		CreatedBy:            &this.CreatedBy,
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
