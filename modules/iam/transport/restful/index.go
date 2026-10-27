package restful

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	v1 "github.com/sky-as-code/nikki-erp/modules/iam/transport/restful/v1"
)

// InitRestfulHandlers registers the REST handlers, then each resource's routes from its own
// routes_<resource>.go file.
func InitRestfulHandlers() error {
	err := deps.Register(
		v1.NewUserRest,
		v1.NewUserCrudRest,
		v1.NewServiceCrudRest,
		v1.NewGroupCrudRest,
		v1.NewOrgCrudRest,
		v1.NewOrgUnitCrudRest,
		v1.NewResourceCrudRest,
		v1.NewActionCrudRest,
		v1.NewEntitlementCrudRest,
		v1.NewRoleCrudRest,
		v1.NewGrantRequestCrudRest,
		v1.NewLoginRest,
		v1.NewPasswordRest,
		v1.NewPermissionRest,
	)
	return stdErr.Join(
		err,
		initMeV1(),
		initUserCrudV1(),
		initServiceCrudV1(),
		initOrgV1(),
		initOrgUnitV1(),
		initGroupV1(),
		initResourceV1(),
		initEntitlementV1(),
		initRoleV1(),
		initGrantRequestV1(),
		initSignInV1(),
		initSettingsV1(),
	)
}
