package user

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

func (this CreateUserCommand) ToDomainModel() *domain.User {
	user := &domain.User{}
	model.MustCopy(this, user)
	return user
}

func (this DeleteUserCommand) ToDomainModel() *domain.User {
	user := &domain.User{}
	user.Id = &this.Id
	user.ScopeRef = this.ScopeRef
	return user
}

func (this UpdateUserCommand) ToDomainModel() *domain.User {
	user := &domain.User{}
	model.MustCopy(this, user)
	return user
}
