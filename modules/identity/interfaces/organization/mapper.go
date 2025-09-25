package organization

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

func (this CreateOrganizationCommand) ToDomainModel() *domain.Organization {
	org := &domain.Organization{}
	model.MustCopy(this, org)

	return org
}

func (this UpdateOrganizationCommand) ToDomainModel() *domain.Organization {
	org := &domain.Organization{}
	model.MustCopy(this, org)

	return org
}
