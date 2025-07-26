package app

import (
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

func InitServices() error {
	err := errors.Join(
		deps.Register(NewAttemptServiceImpl),
		deps.Register(NewLoginServiceImpl),
		deps.Register(NewPasswordServiceImpl),
	)
	return err
}
