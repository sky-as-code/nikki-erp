package repository

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

func InitRepositories() error {
	err := stdErr.Join(
	// orm.RegisterEntity(BuildUserDescriptor()),
	// orm.RegisterEntity(BuildGroupDescriptor()),
	// orm.RegisterEntity(BuildOrganizationDescriptor()),
	// orm.RegisterEntity(BuildOrganizationUnitDescriptor()),
	)
	if err != nil {
		return err
	}

	err = stdErr.Join(
		// deps.Invoke(registerIdentitySearchPredicates),
		deps.Register(NewActionDynamicRepository),
		deps.Register(NewEntitlementDynamicRepository),
		deps.Register(NewGroupDynamicRepository),
		deps.Register(NewOrganizationDynamicRepository),
		deps.Register(NewOrgUnitDynamicRepository),
		deps.Register(NewPermissionDynamicRepository),
		deps.Register(NewResourceDynamicRepository),
		deps.Register(NewRoleDynamicRepository),
		deps.Register(NewRoleRequestDynamicRepository),
		deps.Register(NewUserDynamicRepository),
	)

	return err
}
