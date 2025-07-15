package user

import (
	"reflect"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

func (this CreateUserCommand) ToUser() *domain.User {
	user := &domain.User{}
	model.MustCopy(this, user)
	user.PasswordRaw = &this.Password
	return user
}

func (this UpdateUserCommand) ToUser() *domain.User {
	user := &domain.User{}
	model.MustCopy(this, user)
	user.PasswordRaw = this.Password
	return user
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
