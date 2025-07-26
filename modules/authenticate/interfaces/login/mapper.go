package login

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
)

func (this CreateLoginAttemptCommand) ToLoginAttempt() *domain.LoginAttempt {
	attempt := &domain.LoginAttempt{}
	model.MustCopy(this, attempt)
	return attempt
}

func (this UpdateLoginAttemptCommand) ToLoginAttempt() *domain.LoginAttempt {
	attempt := &domain.LoginAttempt{}
	model.MustCopy(this, attempt)
	return attempt
}
