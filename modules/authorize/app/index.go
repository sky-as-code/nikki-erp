package app

import (
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

func InitServices() error {
	err := errors.Join(
		deps.Register(NewResourceServiceImpl),
		deps.Register(NewActionServiceImpl),
		deps.Register(NewEntitlementServiceImpl),
		deps.Register(NewEntitlementAssignmentServiceImpl),
		deps.Register(NewRoleServiceImpl),
		deps.Register(NewRoleSuiteServiceImpl),
		deps.Register(NewAuthorizeServiceImpl),
	)
	return err
}
