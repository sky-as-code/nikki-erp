package app

import (
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

func InitServices() error {
	err := errors.Join(
		deps.Register(NewActionServiceImpl),
		deps.Register(NewEntitlementServiceImpl),
		deps.Register(NewGroupServiceImpl),
		deps.Register(NewOrganizationServiceImpl),
		deps.Register(NewOrgUnitServiceImpl),
		deps.Register(NewPermissionServiceImpl),
		deps.Register(NewResourceServiceImpl),
		deps.Register(NewRoleServiceImpl),
		deps.Register(NewRoleRequestServiceImpl),
		deps.Register(NewUserServiceImpl),
	)
	return err
}
