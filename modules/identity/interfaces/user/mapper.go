package user

import (
	"reflect"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

func (this CreateUserCommand) ToUser() *domain.User {
	return &domain.User{
		DisplayName:        &this.DisplayName,
		Email:              &this.Email,
		MustChangePassword: &this.MustChangePassword,
		PasswordRaw:        &this.Password,
	}
}

func (this UpdateUserCommand) ToUser() *domain.User {
	return &domain.User{
		ModelBase: model.ModelBase{
			Id:   &this.Id,
			Etag: &this.Etag,
		},
		AvatarUrl:          this.AvatarUrl,
		DisplayName:        this.DisplayName,
		Email:              this.Email,
		MustChangePassword: this.MustChangePassword,
		PasswordRaw:        this.Password,
		StatusValue:        this.Status,
	}
}

func init() {
	model.AddConversion[domain.UserStatus, string](func(in reflect.Value) (reflect.Value, error) {
		status := in.Interface().(domain.UserStatus)
		if status.Value == nil {
			return reflect.ValueOf(""), nil
		}
		return reflect.ValueOf(*status.Value), nil
	})

	model.AddConversion[*domain.UserStatus, string](func(in reflect.Value) (reflect.Value, error) {
		if in.IsNil() {
			return reflect.ValueOf(""), nil
		}

		status := in.Interface().(*domain.UserStatus)
		if status.Value == nil {
			return reflect.ValueOf(""), nil
		}
		return reflect.ValueOf(*status.Value), nil
	})
}
