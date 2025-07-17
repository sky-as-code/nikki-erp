package role

import (
	"github.com/sky-as-code/nikki-erp/common/model"
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
		Entitlements:         this.ToEntitlements(),
	}
}

func (this CreateRoleCommand) ToEntitlements() []*domain.Entitlement {
	entitlements := make([]*domain.Entitlement, 0)
	for _, entitlementId := range this.Entitlements {
		entitlements = append(entitlements, &domain.Entitlement{
			ModelBase: model.ModelBase{
				Id: entitlementId,
			},
		})
	}

	return entitlements
}
