package user

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

func (this CreateUserCommand) ToUser() *domain.User {
	var status *domain.UserStatus
	if this.IsEnabled {
		status = util.ToPtr(domain.UserStatusActive)
	} else {
		status = util.ToPtr(domain.UserStatusInactive)
	}

	return &domain.User{
		DisplayName:        &this.DisplayName,
		Email:              &this.Email,
		MustChangePassword: &this.MustChangePassword,
		PasswordRaw:        &this.Password,
		Status:             status,

		AuditableBase: model.AuditableBase{
			CreatedBy: util.ToPtr(model.Id(this.CreatedBy)),
		},
	}
}

func (this UpdateUserCommand) ToUser() *domain.User {
	var status *domain.UserStatus
	status = nil
	if this.IsEnabled != nil && *this.IsEnabled {
		status = util.ToPtr(domain.UserStatusActive)
	} else if this.IsEnabled != nil {
		status = util.ToPtr(domain.UserStatusInactive)
	}

	return &domain.User{
		ModelBase: model.ModelBase{
			Id: model.WrapId(this.Id),
		},
		AuditableBase: model.AuditableBase{
			UpdatedBy: util.ToPtr(model.Id(this.UpdatedBy)),
		},
		AvatarUrl:          this.AvatarUrl,
		DisplayName:        this.DisplayName,
		Email:              this.Email,
		MustChangePassword: this.MustChangePassword,
		PasswordRaw:        this.Password,
		Status:             status,
	}
}
