package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v4"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver/middlewares"
	"github.com/sky-as-code/nikki-erp/modules/identity/constants"
	v1 "github.com/sky-as-code/nikki-erp/modules/identity/transport/restful/v1"
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
		// v1.NewRoleRequestRest,
	)
	err = stdErr.Join(
		err,
		initIdentityV1(),
		initAuthorizeV1(),
	)
	return err
}

func initIdentityV1() error {
	return deps.Invoke(func(
		route *echo.Group,
		cqrsBus cqrs.CqrsBus,
		userRest *v1.UserRest,
		groupRest *v1.GroupRest,
		orgRest *v1.OrganizationRest,
		orgunitRest *v1.OrgUnitRest,
	) {
		v1 := route.Group("/v1/identity")
		v1.Use(middlewares.RequestContextMiddleware2(constants.IdentityModuleName))
		// protected := route.Group("", commonMiddleware.RequireAuthMiddleware())
		// checker := commonMiddleware.NewCqrsPermissionChecker(cqrsBus)

		// mwUserView := commonMiddleware.RequirePermission(checker, constants.ResourceUser, constants.ActionView, nil)
		// mwUserCreate := commonMiddleware.RequirePermission(checker, constants.ResourceUser, constants.ActionCreate, nil)
		// mwUserUpdate := commonMiddleware.RequirePermission(checker, constants.ResourceUser, constants.ActionUpdate, nil)
		// mwUserDelete := commonMiddleware.RequirePermission(checker, constants.ResourceUser, constants.ActionDelete, nil)

		// mwGroupView := commonMiddleware.RequirePermission(checker, constants.ResourceGroup, constants.ActionView, nil)
		// mwGroupCreate := commonMiddleware.RequirePermission(checker, constants.ResourceGroup, constants.ActionCreate, nil)
		// mwGroupUpdate := commonMiddleware.RequirePermission(checker, constants.ResourceGroup, constants.ActionUpdate, nil)
		// mwGroupDelete := commonMiddleware.RequirePermission(checker, constants.ResourceGroup, constants.ActionDelete, nil)

		// mwOrgView := commonMiddleware.RequirePermission(checker, constants.ResourceOrganization, constants.ActionView, nil)
		// mwOrgCreate := commonMiddleware.RequirePermission(checker, constants.ResourceOrganization, constants.ActionCreate, nil)
		// mwOrgUpdate := commonMiddleware.RequirePermission(checker, constants.ResourceOrganization, constants.ActionUpdate, nil)
		// mwOrgDelete := commonMiddleware.RequirePermission(checker, constants.ResourceOrganization, constants.ActionDelete, nil)

		// mwOrgUnitView := commonMiddleware.RequirePermission(checker, constants.ResourceOrganizationUnit, constants.ActionView, nil)
		// mwOrgUnitCreate := commonMiddleware.RequirePermission(checker, constants.ResourceOrganizationUnit, constants.ActionCreate, nil)
		// mwOrgUnitUpdate := commonMiddleware.RequirePermission(checker, constants.ResourceOrganizationUnit, constants.ActionUpdate, nil)
		// mwOrgUnitDelete := commonMiddleware.RequirePermission(checker, constants.ResourceOrganizationUnit, constants.ActionDelete, nil)
		// mwOrgUnitManageUsers := commonMiddleware.RequirePermission(checker, constants.ResourceOrganizationUnit, constants.ActionManageUsers, nil)

		v1.DELETE("/users/:id", userRest.DeleteUser)
		v1.GET("/users/:id", userRest.GetUser)
		v1.GET("/users", userRest.SearchUsers)
		v1.POST("/users/exists", userRest.UserExists)
		v1.POST("/users/:id/archived", userRest.SetUserIsArchived)
		v1.POST("/users", userRest.CreateUser)
		v1.PUT("/users/:id", userRest.UpdateUser)
		// protected.POST("/users", userRest.CreateUser, middlewares.RequestContextMiddleware2, mwUserCreate)
		// protected.DELETE("/users/:id", userRest.DeleteUser, mwUserDelete)
		// protected.GET("/users/:id", userRest.GetUserById, mwUserView)
		// protected.GET("/users", userRest.SearchUsers, mwUserView)
		// protected.PUT("/users/:id", userRest.UpdateUser, mwUserUpdate)
		// route.GET("/users/context", userRest.GetUserContext)

		// protected.PUT("/users2/:id", userRest.UpdateUser2, mwUserUpdate, middlewares.RequestContextMiddleware2)
		// protected.GET("/users2/:id", userRest.GetUserById2, mwUserView, middlewares.RequestContextMiddleware2)
		// protected.POST("/users2/search", userRest.SearchUsers2, mwUserView, middlewares.RequestContextMiddleware2)
		// protected.POST("/users2/:id/archive", userRest.ArchiveUser2, mwUserUpdate, middlewares.RequestContextMiddleware2)

		v1.DELETE("/groups/:id", groupRest.DeleteGroup)
		v1.GET("/groups/:id", groupRest.GetGroup)
		v1.GET("/groups", groupRest.SearchGroups)
		v1.POST("/groups/exists", groupRest.GroupExists)
		v1.POST("/groups/:group_id/manage-users", groupRest.ManageGroupUsers)
		v1.POST("/groups", groupRest.CreateGroup)
		v1.PUT("/groups/:id", groupRest.UpdateGroup)

		v1.DELETE("/organizations/:id", orgRest.DeleteOrg)
		v1.GET("/organizations/:id", orgRest.GetOrg)
		v1.GET("/organizations", orgRest.SearchOrgs)
		v1.POST("/organizations/:id/archived", orgRest.SetOrgIsArchived)
		v1.POST("/organizations/:org_id/manage-users", orgRest.ManageOrgUsers)
		v1.POST("/organizations/exists", orgRest.OrgExists)
		v1.POST("/organizations", orgRest.CreateOrg)
		v1.PUT("/organizations/:id", orgRest.UpdateOrg)

		v1.DELETE("/orgunits/:id", orgunitRest.DeleteOrgUnit)
		v1.GET("/orgunits/:id", orgunitRest.GetOrgUnit)
		v1.GET("/orgunits", orgunitRest.SearchOrgUnits)
		v1.POST("/orgunits/:id/exists", orgunitRest.OrgUnitExists)
		v1.POST("/orgunits/:orgunit_id/manage-users", orgunitRest.ManageOrgUnitUsers)
		v1.POST("/orgunits", orgunitRest.CreateOrgUnit)
		v1.PUT("/orgunits/:id", orgunitRest.UpdateOrgUnit)
	})
}

func initAuthorizeV1() error {
	return deps.Invoke(func(
		route *echo.Group,
		cqrsBus cqrs.CqrsBus,
		resourceRest *v1.ResourceRest,
		actionRest *v1.ActionRest,
		entitlementRest *v1.EntitlementRest,
		roleRest *v1.RoleRest,
		// roleRequestRest *v1.RoleRequestRest,
	) {
		v1 := route.Group("/v1/authorize")
		v1.Use(middlewares.RequestContextMiddleware2(constants.IdentityModuleName))

		v1.DELETE("/resources/:resource_id/actions/:action_id", actionRest.DeleteAction)
		v1.GET("/resources/:resource_id/actions/:action_id", actionRest.GetAction)
		v1.GET("/resources/:resource_id/actions", actionRest.SearchActions)
		v1.POST("/resources/:resource_id/actions/exists", actionRest.ActionExists)
		v1.POST("/resources/:resource_id/actions", actionRest.CreateAction)
		v1.PUT("/resources/:resource_id/actions/:action_id", actionRest.UpdateAction)

		v1.DELETE("/resources/:id", resourceRest.DeleteResource)
		v1.GET("/resources/:id", resourceRest.GetResource)
		v1.GET("/resources", resourceRest.SearchResources)
		v1.POST("/resources/exists", resourceRest.ResourceExists)
		v1.POST("/resources", resourceRest.CreateResource)
		v1.PUT("/resources/:id", resourceRest.UpdateResource)

		v1.DELETE("/entitlements/:id", entitlementRest.DeleteEntitlement)
		v1.GET("/entitlements/:id", entitlementRest.GetEntitlement)
		v1.GET("/entitlements", entitlementRest.SearchEntitlements)
		v1.POST("/entitlements/:entitlement_id/manage-roles", entitlementRest.ManageEntitlementRoles)
		v1.POST("/entitlements/:id/archived", entitlementRest.SetEntitlementIsArchived)
		v1.POST("/entitlements/exists", entitlementRest.EntitlementExists)
		v1.POST("/entitlements", entitlementRest.CreateEntitlement)
		v1.PUT("/entitlements/:id", entitlementRest.UpdateEntitlement)

		v1.DELETE("/roles/:id", roleRest.DeleteRole)
		v1.GET("/roles/:id", roleRest.GetRole)
		v1.GET("/roles", roleRest.SearchRoles)
		v1.POST("/roles/:role_id/manage-entitlements", roleRest.ManageRoleEntitlements)
		v1.POST("/roles/:id/archived", roleRest.SetRoleIsArchived)
		v1.POST("/roles/exists", roleRest.RoleExists)
		v1.POST("/roles", roleRest.CreateRole)
		v1.PUT("/roles/:id", roleRest.UpdateRole)

		// v1.DELETE("/grant-requests/:id", roleRequestRest.DeleteRoleRequest)
		// v1.GET("/grant-requests/:id", roleRequestRest.GetRoleRequest)
		// v1.GET("/grant-requests", roleRequestRest.SearchRoleRequests)
		// v1.POST("/grant-requests/exists", roleRequestRest.RoleRequestExists)
		// v1.POST("/grant-requests", roleRequestRest.CreateRoleRequest)
		// v1.PUT("/grant-requests/:id", roleRequestRest.UpdateRoleRequest)
	})
}
