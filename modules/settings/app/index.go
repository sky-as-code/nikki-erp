package app

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

func InitApplicationServices() error {
	return stdErr.Join(
		deps.Register(NewUserPreferenceApplicationServiceImpl),
	)
}
