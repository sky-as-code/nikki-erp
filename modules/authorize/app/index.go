package app

import (
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

func InitServices() error {
	err := errors.Join(
		deps.Register(NewActionServiceImpl),
		deps.Register(NewAuthorizeServiceImpl),
		deps.Register(NewEntitlementServiceImpl),
		deps.Register(NewEntitlementAssignmentServiceImpl),
		deps.Register(NewGrantRequestServiceImpl),
		deps.Register(NewResourceServiceImpl),
		deps.Register(NewRoleServiceImpl),
		deps.Register(NewRoleSuiteServiceImpl),
	)
	return err
}
