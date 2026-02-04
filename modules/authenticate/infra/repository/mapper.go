package repository

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/infra/ent"
)

func entToAttempt(dbAttempt *ent.LoginAttempt) *domain.LoginAttempt {
	attempt := &domain.LoginAttempt{}
	model.MustCopy(dbAttempt, attempt)

	attempt.DeviceIp = dbAttempt.DeviceIP
	return attempt
}

func entToPasswordStore(dbPasswordStore *ent.PasswordStore) *domain.PasswordStore {
	passStore := &domain.PasswordStore{}
	model.MustCopy(dbPasswordStore, passStore)

	// Manualy copy because sensitive fields are not iteratable.
	// passStore.Password = dbPasswordStore.Password
	// passStore.Passwordotp = dbPasswordStore.Passwordotp
	// passStore.PasswordotpRecovery = dbPasswordStore.PasswordotpRecovery
	// passStore.Passwordtmp = dbPasswordStore.Passwordtmp

	return passStore
}
