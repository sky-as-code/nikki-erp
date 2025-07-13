package organization

import (
	"reflect"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

func (this CreateOrganizationCommand) ToOrganization() *domain.Organization {
	org := &domain.Organization{}
	model.MustCopy(this, org)

	return org
}

func (this UpdateOrganizationCommand) ToOrganization() *domain.Organization {
	org := &domain.Organization{}
	model.MustCopy(this, org)
	org.StatusValue = this.Status

	return org
}

func init() {
	model.AddConversion[domain.IdentityStatus, string](func(in reflect.Value) (reflect.Value, error) {
		result := in.Interface().(domain.IdentityStatus)
		if result.Value == nil {
			return reflect.ValueOf(""), nil
		}
		return reflect.ValueOf(*result.Value), nil
	})

	model.AddConversion[domain.IdentityStatus, *string](func(in reflect.Value) (reflect.Value, error) {
		result := in.Interface().(domain.IdentityStatus)
		return reflect.ValueOf(result.Value), nil
	})

	model.AddConversion[*domain.IdentityStatus, string](func(in reflect.Value) (reflect.Value, error) {
		if in.IsNil() {
			return reflect.ValueOf(""), nil
		}
		result := in.Interface().(*domain.IdentityStatus)
		if result.Value == nil {
			return reflect.ValueOf(""), nil
		}
		return reflect.ValueOf(*result.Value), nil
	})

	model.AddConversion[*domain.IdentityStatus, *string](func(in reflect.Value) (reflect.Value, error) {
		if in.IsNil() {
			return reflect.ValueOf((*string)(nil)), nil
		}
		result := in.Interface().(*domain.IdentityStatus)
		return reflect.ValueOf(result.Value), nil
	})
}
