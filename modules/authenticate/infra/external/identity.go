package external

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	itExt "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/external"
	itGrp "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/group"
	itOrg "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
	itOrgUnit "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/orgunit"
	itUsr "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

func InitExternal() error {

	// This will be replaced with the actual implementation when this application is
	// split into separate microservices.
	err := stdErr.Join(
		deps.Register(func(orgSvc itOrg.OrganizationDomainService) itExt.OrganizationExtService {
			return orgSvc
		}),
		deps.Register(func(orgUnitSvc itOrgUnit.OrgUnitDomainService) itExt.OrgUnitExtService {
			return orgUnitSvc
		}),
		deps.Register(func(groupSvc itGrp.GroupDomainService) itExt.GroupExtService {
			return groupSvc
		}),
		deps.Register(func(userSvc itUsr.UserDomainService) itExt.UserExtService {
			return userSvc
		}),
	)

	return err
}
