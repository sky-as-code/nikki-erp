package app

import (
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

func InitServices() error {
	err := errors.Join(
		deps.Register(NewUserServiceImpl),
		deps.Register(NewGroupServiceImpl),
		deps.Register(NewOrganizationServiceImpl),
	)
	return err
}
