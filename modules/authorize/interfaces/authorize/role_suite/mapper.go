package role_suite

import (
	"github.com/sky-as-code/nikki-erp/common/model"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

func (this CreateRoleSuiteCommand) ToRoleSuite() *domain.RoleSuite {
	return &domain.RoleSuite{
		Name:                 &this.Name,
		Description:          this.Description,
		OwnerType:            domain.WrapRoleSuiteOwnerType(this.OwnerType),
		OwnerRef:             &this.OwnerRef,
		IsRequestable:        this.IsRequestable,
		IsRequiredAttachment: this.IsRequiredAttachment,
		IsRequiredComment:    this.IsRequiredComment,
		CreatedBy:            &this.CreatedBy,
		Roles:                this.ToRoles(),
	}
}

func (this CreateRoleSuiteCommand) ToRoles() []domain.Role {
	roles := make([]domain.Role, 0)
	for _, roleId := range this.Roles {
		roles = append(roles, domain.Role{
			ModelBase: model.ModelBase{
				Id: roleId,
			},
		})
	}

	return roles
}
