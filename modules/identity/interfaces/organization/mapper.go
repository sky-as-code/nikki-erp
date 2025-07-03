package organization

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

func (this CreateOrganizationCommand) ToOrganization() *domain.Organization {
	return &domain.Organization{
		DisplayName: this.DisplayName,
		Address:     this.Address,
		LegalName:   this.LegalName,
		PhoneNumber: this.PhoneNumber,
		Slug:        &this.Slug,
		Status:      this.Status,
	}
}

func (this UpdateOrganizationCommand) ToOrganization() *domain.Organization {
	return &domain.Organization{
		ModelBase: model.ModelBase{
			Etag: &this.Etag,
		},
		DisplayName: this.DisplayName,
		Address:     this.Address,
		LegalName:   this.LegalName,
		PhoneNumber: this.PhoneNumber,
		Slug:        &this.Slug,
		Status:      this.Status,
	}
}
