package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	m "github.com/sky-as-code/nikki-erp/modules/core/httpserver/middlewares"
	d "github.com/sky-as-code/nikki-erp/modules/identity/domain"
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
	) error {
		routeV1 := route.Group("/v1/identity")

		routeV1.DELETE("/users/:id", userRest.DeleteUser, m.Authorized(d.UserActionDelete, d.UserResourceCode, d.UserAuthScope))
		routeV1.GET("/users/:id", userRest.GetUser, m.Authorized(d.UserActionView, d.UserResourceCode, d.UserAuthScope))
		routeV1.GET("/users", userRest.SearchUsers, m.Authorized(d.UserActionView, d.UserResourceCode, d.UserAuthScope))
		routeV1.POST("/users/exists", userRest.UserExists, m.Authorized(d.UserActionView, d.UserResourceCode, d.UserAuthScope))
		routeV1.POST("/users/:id/archived", userRest.SetUserIsArchived, m.Authorized(d.UserActionSetArchived, d.UserResourceCode, d.UserAuthScope))
		routeV1.POST("/users", userRest.CreateUser, m.Authorized(d.UserActionCreate, d.UserResourceCode, d.UserAuthScope))
		routeV1.PUT("/users/:id", userRest.UpdateUser, m.Authorized(d.UserActionUpdate, d.UserResourceCode, d.UserAuthScope))

		routeV1.DELETE("/groups/:id", groupRest.DeleteGroup, m.Authorized(d.GroupActionDelete, d.GroupResourceCode, d.GroupAuthScope))
		routeV1.GET("/groups/:id", groupRest.GetGroup, m.Authorized(d.GroupActionView, d.GroupResourceCode, d.GroupAuthScope))
		routeV1.GET("/groups", groupRest.SearchGroups, m.Authorized(d.GroupActionView, d.GroupResourceCode, d.GroupAuthScope))
		routeV1.POST("/groups/exists", groupRest.GroupExists, m.Authorized(d.GroupActionView, d.GroupResourceCode, d.GroupAuthScope))
		routeV1.POST("/groups/:group_id/manage-users", groupRest.ManageGroupUsers, m.Authorized(d.GroupActionManageUsers, d.GroupResourceCode, d.GroupAuthScope))
		routeV1.POST("/groups", groupRest.CreateGroup, m.Authorized(d.GroupActionCreate, d.GroupResourceCode, d.GroupAuthScope))
		routeV1.PUT("/groups/:id", groupRest.UpdateGroup, m.Authorized(d.GroupActionUpdate, d.GroupResourceCode, d.GroupAuthScope))

		routeV1.DELETE("/organizations/:id", orgRest.DeleteOrg, m.Authorized(d.OrgActionDelete, d.OrgResourceCode, d.OrgAuthScope))
		routeV1.GET("/organizations/:id", orgRest.GetOrg, m.Authorized(d.OrgActionView, d.OrgResourceCode, d.OrgAuthScope))
		routeV1.GET("/organizations", orgRest.SearchOrgs, m.Authorized(d.OrgActionView, d.OrgResourceCode, d.OrgAuthScope))
		routeV1.POST("/organizations/exists", orgRest.OrgExists, m.Authorized(d.OrgActionView, d.OrgResourceCode, d.OrgAuthScope))
		routeV1.POST("/organizations/:id/archived", orgRest.SetOrgIsArchived, m.Authorized(d.OrgActionSetArchived, d.OrgResourceCode, d.OrgAuthScope))
		routeV1.POST("/organizations/:org_id/manage-users", orgRest.ManageOrgUsers, m.Authorized(d.OrgActionManageUsers, d.OrgResourceCode, d.OrgAuthScope))
		routeV1.POST("/organizations", orgRest.CreateOrg, m.Authorized(d.OrgActionCreate, d.OrgResourceCode, d.OrgAuthScope))
		routeV1.PUT("/organizations/:id", orgRest.UpdateOrg, m.Authorized(d.OrgActionUpdate, d.OrgResourceCode, d.OrgAuthScope))

		routeV1.DELETE("/orgunits/:id", orgunitRest.DeleteOrgUnit, m.Authorized(d.OrgUnitActionDelete, d.OrgUnitResourceCode, d.OrgUnitAuthScope))
		routeV1.GET("/orgunits/:id", orgunitRest.GetOrgUnit, m.Authorized(d.OrgUnitActionView, d.OrgUnitResourceCode, d.OrgUnitAuthScope))
		routeV1.GET("/orgunits", orgunitRest.SearchOrgUnits, m.Authorized(d.OrgUnitActionView, d.OrgUnitResourceCode, d.OrgUnitAuthScope))
		routeV1.POST("/orgunits/:id/exists", orgunitRest.OrgUnitExists, m.Authorized(d.OrgUnitActionView, d.OrgUnitResourceCode, d.OrgUnitAuthScope))
		routeV1.POST("/orgunits/:orgunit_id/manage-users", orgunitRest.ManageOrgUnitUsers, m.Authorized(d.OrgUnitActionManageUsers, d.OrgUnitResourceCode, d.OrgUnitAuthScope))
		routeV1.POST("/orgunits", orgunitRest.CreateOrgUnit, m.Authorized(d.OrgUnitActionCreate, d.OrgUnitResourceCode, d.OrgUnitAuthScope))
		routeV1.PUT("/orgunits/:id", orgunitRest.UpdateOrgUnit, m.Authorized(d.OrgUnitActionUpdate, d.OrgUnitResourceCode, d.OrgUnitAuthScope))

		return nil
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
