package services

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

func InitDomainServices() error {
	return stdErr.Join(
		deps.Register(NewUserPreferenceCrudDomainServiceImpl),
		deps.Register(NewUserPreferenceUiDomainServiceImpl),
	)
}
