package services

import deps "github.com/sky-as-code/nikki-erp/common/deps_inject"

func InitDomainServices() error {
	return deps.Register(
		NewAttemptDomainServiceImpl,
		NewLoginDomainServiceImpl,
		NewPasswordDomainServiceImpl,
	)
}
