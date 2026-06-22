package external

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	itExt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/external"

	// itGrp "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/group"
	// itOrg "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/organization"
	// itOrgUnit "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/orgunit"
	// itUsr "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
	itSet "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/userpref"
)

func InitExternalServices() error {
	return stdErr.Join(
		deps.Register(func(userPrefSvc itSet.UserPreferenceUiDomainService) itExt.UserPreferenceUiDomainService {
			return userPrefSvc
		}),
		// deps.Register(func(orgSvc itOrg.OrganizationDomainService) itExt.OrganizationExtService {
		// 	return orgSvc
		// }),
		// deps.Register(func(orgUnitSvc itOrgUnit.OrgUnitDomainService) itExt.OrgUnitExtService {
		// 	return orgUnitSvc
		// }),
		// deps.Register(func(groupSvc itGrp.GroupDomainService) itExt.GroupExtService {
		// 	return groupSvc
		// }),
		// deps.Register(func(userSvc itUsr.UserDomainService) itExt.UserExtService {
		// 	return userSvc
		// }),
	)
}
