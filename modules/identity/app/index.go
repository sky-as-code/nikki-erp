package app

import (
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

func InitApplicationServices() error {
	err := errors.Join(
		deps.Register(NewEntitlementApplicationServiceImpl),
		deps.Register(NewGroupApplicationServiceImpl),
		deps.Register(NewOrganizationApplicationServiceImpl),
		deps.Register(NewOrgUnitApplicationServiceImpl),
		deps.Register(NewResourceApplicationServiceImpl),
		deps.Register(NewActionApplicationService),
		deps.Register(NewRoleApplicationServiceImpl),
		deps.Register(NewRoleRequestApplicationServiceImpl),
		deps.Register(NewPermissionApplicationServiceImpl),
		deps.Register(NewUserApplicationServiceImpl),
	)
	return err
}
