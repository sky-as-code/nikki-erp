package user

import (
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
