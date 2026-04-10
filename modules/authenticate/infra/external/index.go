package external

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	itExt "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/external"
	"github.com/sky-as-code/nikki-erp/modules/identity/app"
	itGrp "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/group"
	itOrg "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
	itOrgUnit "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/orgunit"
	itUsr "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

// This will be replaced with the actual implementation when this application is
// split into separate microservices.
type OrganizationExtServiceImpl = app.OrganizationServiceImpl
type OrgUnitExtServiceImpl = app.OrgUnitServiceImpl
type GroupExtServiceImpl = app.GroupServiceImpl
type UserExtServiceImpl = app.UserServiceImpl

func InitExternal() error {
	err := stdErr.Join(
		deps.Register(func(orgSvc itOrg.OrganizationService) itExt.OrganizationExtService {
			return orgSvc
		}),
		deps.Register(func(orgUnitSvc itOrgUnit.OrgUnitService) itExt.OrgUnitExtService {
			return orgUnitSvc
		}),
		deps.Register(func(groupSvc itGrp.GroupService) itExt.GroupExtService {
			return groupSvc
		}),
		deps.Register(func(userSvc itUsr.UserService) itExt.UserExtService {
			return userSvc
		}),
	)

	return err
}
