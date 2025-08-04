package role

import (
	"github.com/sky-as-code/nikki-erp/common/model"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

func (this CreateRoleCommand) ToRole() *domain.Role {
	role := &domain.Role{}
	model.MustCopy(this, role)
	return role
}

func (this UpdateRoleCommand) ToRole() *domain.Role {
	role := &domain.Role{}
	model.MustCopy(this, role)
	return role
}
