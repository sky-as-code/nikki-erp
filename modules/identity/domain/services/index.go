package services

import (
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

func InitDomainServices() error {
	err := errors.Join(
		deps.Register(NewEntitlementDomainServiceImpl),
		deps.Register(NewGroupDomainServiceImpl),
		deps.Register(NewOrganizationDomainServiceImpl),
		deps.Register(NewOrgUnitDomainServiceImpl),
		deps.Register(NewPermissionDomainServiceImpl),
		deps.Register(NewResourceDomainServiceImpl),
		deps.Register(NewActionDomainService),
		deps.Register(NewRoleDomainServiceImpl),
		deps.Register(NewRoleRequestDomainServiceImpl),
		deps.Register(NewUserDomainServiceImpl),
	)
	return err
}
