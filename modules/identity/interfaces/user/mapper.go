package user

import (
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
	user.StatusValue = this.Status
	return user
}
