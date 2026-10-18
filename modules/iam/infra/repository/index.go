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
		deps.Register(NewPermissionDynamicRepository),
		deps.Register(NewPermissionHistoryDynamicRepository),
		deps.Register(NewAttemptDynamicRepository),
		deps.Register(NewPasswordStoreDynamicRepository),
	)

	return err
}
