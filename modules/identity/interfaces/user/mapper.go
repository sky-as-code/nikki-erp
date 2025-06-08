package user

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

func (this CreateUserCommand) ToUser() *domain.User {
	var status *domain.UserStatus
	if this.IsActive {
		status = util.ToPtr(domain.UserStatusActive)
	}

	return &domain.User{
		DisplayName:        &this.DisplayName,
		Email:              &this.Email,
		MustChangePassword: &this.MustChangePassword,
		PasswordRaw:        &this.Password,
		Status:             status,
	}
}

func (this UpdateUserCommand) ToUser() *domain.User {
	var status *domain.UserStatus
	status = nil
	if this.IsActive != nil && *this.IsActive {
		status = util.ToPtr(domain.UserStatusActive)
	} else if this.IsActive != nil {
		status = util.ToPtr(domain.UserStatusInactive)
	}

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
		Status:             status,
	}
}
