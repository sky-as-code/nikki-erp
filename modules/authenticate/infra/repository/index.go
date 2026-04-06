package repository

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

func InitRepositories() error {
	err := stdErr.Join(
		deps.Register(NewAttemptDynamicRepository),
		deps.Register(NewPasswordStoreDynamicRepository),
	)
	return err
}
