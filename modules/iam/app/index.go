package app

import (
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

func InitApplicationServices() error {
	err := errors.Join(
		deps.Register(NewSettingsApplicationServiceImpl),
		deps.Register(NewAttemptApplicationServiceImpl),
		deps.Register(NewLoginApplicationServiceImpl),
		deps.Register(NewPasswordApplicationServiceImpl),
		deps.Register(NewPrincipalRolesApplicationService),
		deps.Register(NewPermissionApplicationServiceImpl),
	)
	return err
}
