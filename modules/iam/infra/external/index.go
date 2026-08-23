package external

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	itExt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/external"
	itSettings "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"

	// itGrp "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/group"
	// itOrg "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/organization"
	// itOrgUnit "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/orgunit"
	// itUsr "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
)

// InitExternalServices binds iam's ports to their providing modules.
//
// This is the ONLY package in iam that may import another module. Everything else depends on the
// interfaces in interfaces/external, so a module split changes this file and nothing else.
func InitExternalServices() error {
	return stdErr.Join(
		deps.Register(func(orgSettingsSvc itSettings.OrgSettingsAppService) itExt.OrgSettingsInitExtService {
			return orgSettingsSvc
		}),
		deps.Register(func(userPrefsSvc itSettings.UserPreferencesAppService) itExt.UserSettingsExtService {
			return userPrefsSvc
		}),
		deps.Register(func(tenantSettingsSvc itSettings.TenantSettingsAppService) itExt.SettingsRegistrationExtService {
			return tenantSettingsSvc
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
