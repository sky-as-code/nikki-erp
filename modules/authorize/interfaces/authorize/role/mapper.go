package role

import (
	"github.com/sky-as-code/nikki-erp/common/model"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

func (this CreateRoleCommand) ToDomainModel() *domain.Role {
	role := &domain.Role{}
	model.MustCopy(this, role)
	return role
}

func (this UpdateRoleCommand) ToDomainModel() *domain.Role {
	role := &domain.Role{}
	model.MustCopy(this, role)
	return role
}

func (this DeleteRoleHardCommand) ToDomainModel() *domain.Role {
	role := &domain.Role{}
	model.MustCopy(this, role)
	return role
}

func (this AddEntitlementsCommand) ToDomainModel() *domain.Role {
	role := &domain.Role{}
	model.MustCopy(this, role)
	return role
}

func (this RemoveEntitlementsCommand) ToDomainModel() *domain.Role {
	role := &domain.Role{}
	model.MustCopy(this, role)
	return role
}
