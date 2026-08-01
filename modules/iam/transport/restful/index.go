package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	m "github.com/sky-as-code/nikki-erp/modules/core/httpserver/middlewares"
	v1 "github.com/sky-as-code/nikki-erp/modules/iam/transport/restful/v1"
)

func InitRestfulHandlers() error {
	err := deps.Register(
		v1.NewUserRest,
		v1.NewGroupRest,
		v1.NewOrganizationRest,
		v1.NewOrgUnitRest,
		v1.NewResourceRest,
		v1.NewActionRest,
		v1.NewEntitlementRest,
		v1.NewRoleRest,
		v1.NewLoginRest,
		v1.NewPasswordRest,
		// v1.NewRoleRequestRest,
	)
	err = stdErr.Join(
		err,
		initIamDirectoryV1(),
		initIamAuthorizationV1(),
		initIamSignInV1(),
	)
	return err
}

func initIamDirectoryV1() error {
	return deps.Invoke(func(
		route *echo.Group,
		cqrsBus cqrs.CqrsBus,
		userRest *v1.UserRest,
		groupRest *v1.GroupRest,
		orgRest *v1.OrganizationRest,
		orgunitRest *v1.OrgUnitRest,
	) error {
		routeV1 := route.Group("/v1/iam")

		routeV1.DELETE("/groups/:id", groupRest.DeleteGroup, m.SmokeAuthz())
		routeV1.GET("/groups/meta/schema", groupRest.GetModelSchema, m.SmokeAuthz())
		routeV1.GET("/groups/:id", groupRest.GetGroup, m.SmokeAuthz())
		routeV1.GET("/groups", groupRest.SearchGroups, m.SmokeAuthz())
		routeV1.POST("/groups/exists", groupRest.GroupExists, m.SmokeAuthz())
		routeV1.POST("/groups/:group_id/roles", groupRest.ManageGroupRoleAssignments, m.SmokeAuthz())
		routeV1.POST("/groups/:group_id/manage-users", groupRest.ManageGroupUsers, m.SmokeAuthz())
		routeV1.POST("/groups", groupRest.CreateGroup, m.SmokeAuthz())
		routeV1.PATCH("/groups/:id", groupRest.UpdateGroup, m.SmokeAuthz())

		routeV1.DELETE("/organizations/:id", orgRest.DeleteOrg, m.SmokeAuthz())
		routeV1.GET("/organizations/meta/schema", orgRest.GetModelSchema, m.SmokeAuthz())
		routeV1.GET("/organizations/:id", orgRest.GetOrg, m.SmokeAuthz())
		routeV1.GET("/organizations", orgRest.SearchOrgs, m.SmokeAuthz())
		routeV1.POST("/organizations/exists", orgRest.OrgExists, m.SmokeAuthz())
		routeV1.POST("/organizations/:id/archived", orgRest.SetOrgIsArchived, m.SmokeAuthz())
		routeV1.POST("/organizations/:org_id/manage-users", orgRest.ManageOrgUsers, m.SmokeAuthz())
		routeV1.POST("/organizations", orgRest.CreateOrg, m.SmokeAuthz())
		routeV1.PATCH("/organizations/:id", orgRest.UpdateOrg, m.SmokeAuthz())

		routeV1.DELETE("/orgunits/:id", orgunitRest.DeleteOrgUnit, m.SmokeAuthz())
		routeV1.GET("/orgunits/meta/schema", orgunitRest.GetModelSchema, m.SmokeAuthz())
		routeV1.GET("/orgunits/:id", orgunitRest.GetOrgUnit, m.SmokeAuthz())
		routeV1.GET("/orgunits", orgunitRest.SearchOrgUnits, m.SmokeAuthz())
		routeV1.POST("/orgunits/:id/exists", orgunitRest.OrgUnitExists, m.SmokeAuthz())
		routeV1.POST("/orgunits/:orgunit_id/manage-users", orgunitRest.ManageOrgUnitUsers, m.SmokeAuthz())
		routeV1.POST("/orgunits", orgunitRest.CreateOrgUnit, m.SmokeAuthz())
		routeV1.PATCH("/orgunits/:id", orgunitRest.UpdateOrgUnit, m.SmokeAuthz())

		routeV1.GET("/me/context", userRest.GetUserContext, m.SmokeAuthz())

		routeV1.DELETE("/users/:id", userRest.DeleteUser, m.SmokeAuthz())
		routeV1.GET("/users/meta/schema", userRest.GetModelSchema, m.SmokeAuthz())
		// routeV1.GET("/users/getOne", userRest.GetUser, m.SmokeAuthz())
		routeV1.GET("/users/:id", userRest.GetUser, m.SmokeAuthz())
		routeV1.GET("/users", userRest.SearchUsers, m.SmokeAuthz())
		routeV1.POST("/users/exists", userRest.UserExists, m.SmokeAuthz())
		routeV1.POST("/users/:id/archived", userRest.SetUserIsArchived, m.SmokeAuthz())
		routeV1.POST("/users/:user_id/roles", userRest.ManageUserRoleAssignments, m.SmokeAuthz())
		routeV1.POST("/users", userRest.CreateUser, m.SmokeAuthz())
		// JSON Merge Patch (RFC 7396) semantics.
		routeV1.PATCH("/users/:id", userRest.UpdateUser, m.SmokeAuthz())

		return nil
	})
}

func initIamAuthorizationV1() error {
	return deps.Invoke(func(
		route *echo.Group,
		cqrsBus cqrs.CqrsBus,
		resourceRest *v1.ResourceRest,
		actionRest *v1.ActionRest,
		entitlementRest *v1.EntitlementRest,
		roleRest *v1.RoleRest,
		// roleRequestRest *v1.RoleRequestRest,
	) {
		v1 := route.Group("/v1/iam")

		v1.DELETE("/resources/:resource_id/actions/:action_id", actionRest.DeleteAction, m.SmokeAuthz())
		v1.GET("/resources/actions/schema", resourceRest.GetModelSchema, m.SmokeAuthz())
		v1.GET("/resources/:resource_id/actions/:action_id", actionRest.GetAction, m.SmokeAuthz())
		v1.GET("/resources/:resource_id/actions", actionRest.SearchActions, m.SmokeAuthz())
		v1.POST("/resources/:resource_id/actions/exists", actionRest.ActionExists, m.SmokeAuthz())
		v1.POST("/resources/:resource_id/actions", actionRest.CreateAction, m.SmokeAuthz())
		v1.PUT("/resources/:resource_id/actions/:action_id", actionRest.UpdateAction, m.SmokeAuthz())

		v1.DELETE("/resources/:id", resourceRest.DeleteResource, m.SmokeAuthz())
		v1.GET("/resources/schema", resourceRest.GetModelSchema, m.SmokeAuthz())
		v1.GET("/resources/:id", resourceRest.GetResource, m.SmokeAuthz())
		v1.GET("/resources", resourceRest.SearchResources, m.SmokeAuthz())
		v1.POST("/resources/exists", resourceRest.ResourceExists, m.SmokeAuthz())
		v1.POST("/resources", resourceRest.CreateResource, m.SmokeAuthz())
		v1.PUT("/resources/:id", resourceRest.UpdateResource, m.SmokeAuthz())

		v1.DELETE("/entitlements/:id", entitlementRest.DeleteEntitlement, m.SmokeAuthz())
		v1.GET("/entitlements/schema", entitlementRest.GetModelSchema, m.SmokeAuthz())
		v1.GET("/entitlements/:id", entitlementRest.GetEntitlement, m.SmokeAuthz())
		v1.GET("/entitlements", entitlementRest.SearchEntitlements, m.SmokeAuthz())
		v1.POST("/entitlements/:entitlement_id/manage-roles", entitlementRest.ManageEntitlementRoles, m.SmokeAuthz())
		v1.POST("/entitlements/:id/archived", entitlementRest.SetEntitlementIsArchived, m.SmokeAuthz())
		v1.POST("/entitlements/exists", entitlementRest.EntitlementExists, m.SmokeAuthz())
		v1.POST("/entitlements", entitlementRest.CreateEntitlement, m.SmokeAuthz())
		v1.PUT("/entitlements/:id", entitlementRest.UpdateEntitlement, m.SmokeAuthz())

		// Roles assigned to one principal. Registered here rather than in the directory block
		// because they read the role schema and need roleRest, the same way
		// GET /resources/:resource_id/actions is served by actionRest.
		v1.GET("/users/:user_id/roles", roleRest.SearchUserRoles, m.SmokeAuthz())
		v1.GET("/groups/:group_id/roles", roleRest.SearchGroupRoles, m.SmokeAuthz())

		v1.DELETE("/roles/:id", roleRest.DeleteRole, m.SmokeAuthz())
		v1.GET("/roles/meta/schema", roleRest.GetModelSchema, m.SmokeAuthz())
		// Must precede /roles/:id so "describe" is not bound as an id.
		v1.GET("/roles/describe", roleRest.DescribeRoles, m.SmokeAuthz())
		v1.GET("/roles/:id", roleRest.GetRole, m.SmokeAuthz())
		v1.GET("/roles", roleRest.SearchRoles, m.SmokeAuthz())
		v1.PATCH("/roles/:id", roleRest.UpdateRole, m.SmokeAuthz())
		v1.POST("/roles/:role_id/manage-entitlements", roleRest.ManageRoleEntitlements, m.SmokeAuthz())
		v1.POST("/roles/:id/archived", roleRest.SetRoleIsArchived, m.SmokeAuthz())
		v1.POST("/roles/exists", roleRest.RoleExists, m.SmokeAuthz())
		v1.POST("/roles", roleRest.CreateRole, m.SmokeAuthz())

		// Deprecated aliases kept for existing clients; the canonical shapes above match
		// the shared frontend RestApi client and the directory resources. Remove once no
		// client depends on them.
		v1.GET("/roles/schema", roleRest.GetModelSchema, m.SmokeAuthz())
		v1.PUT("/roles/:id", roleRest.UpdateRole, m.SmokeAuthz())

		// v1.DELETE("/grant-requests/:id", roleRequestRest.DeleteRoleRequest)
		// v1.GET("/grant-requests/:id", roleRequestRest.GetRoleRequest)
		// v1.GET("/grant-requests", roleRequestRest.SearchRoleRequests)
		// v1.POST("/grant-requests/exists", roleRequestRest.RoleRequestExists)
		// v1.POST("/grant-requests", roleRequestRest.CreateRoleRequest)
		// v1.PUT("/grant-requests/:id", roleRequestRest.UpdateRoleRequest)
	})
}

func initIamSignInV1() error {
	return deps.Invoke(func(
		route *echo.Group,
		loginRest *v1.LoginRest,
		passwordRest *v1.PasswordRest,
	) {
		routeV1 := route.Group("/v1/iam")

		routeV1.POST("/signin/start", loginRest.StartSignInFlow, m.PublicUnauthorized)
		routeV1.POST("/signin/continue", loginRest.ContinueSignInFlow, m.PublicUnauthorized)
		routeV1.POST("/signin/refresh", loginRest.RefreshToken, m.PublicUnauthorized)

		routeV1.POST("/passwords/password", passwordRest.SetPassword, m.SmokeAuthz())
		routeV1.POST("/passwords/passwordtmp", passwordRest.CreatePasswordTemp, m.SmokeAuthz())
		routeV1.POST("/passwords/passwordotp", passwordRest.CreatePasswordOtp, m.SmokeAuthz())
		routeV1.POST("/passwords/passwordotp/confirm", passwordRest.ConfirmPasswordOtp, m.SmokeAuthz())
	})
}
