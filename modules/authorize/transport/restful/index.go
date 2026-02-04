package restful

import (
	"errors"

	"github.com/labstack/echo/v4"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	commonMiddleware "github.com/sky-as-code/nikki-erp/common/middleware"
	constants "github.com/sky-as-code/nikki-erp/modules/authorize/constants"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces"
	authzMiddleware "github.com/sky-as-code/nikki-erp/modules/authorize/middleware"
	v1 "github.com/sky-as-code/nikki-erp/modules/authorize/transport/restful/v1"
)

func InitRestfulHandlers() error {
	err := errors.Join(
		initAuthorizeRest(),
	)
	return err
}

func initAuthorizeRest() error {
	deps.Register(
		v1.NewActionRest,
		v1.NewAuthorizeRest,
		v1.NewEntitlementRest,
		v1.NewGrantRequestRest,
		v1.NewResourceRest,
		v1.NewRevokeRequestRest,
		v1.NewRoleRest,
		v1.NewRoleSuiteRest,
	)
	return deps.Invoke(
		func(
			route *echo.Group,
			authzSvc it.AuthorizeService,
			actionRest *v1.ActionRest,
			authorizedRest *v1.AuthorizeRest,
			entitlementRest *v1.EntitlementRest,
			grantRequestRest *v1.GrantRequestRest,
			resourceRest *v1.ResourceRest,
			revokeRequestRest *v1.RevokeRequestRest,
			roleRest *v1.RoleRest,
			roleSuiteRest *v1.RoleSuiteRest,
		) {
			v1 := route.Group("/v1/authorize")
			initV1(v1, authzSvc, actionRest, authorizedRest, entitlementRest, grantRequestRest, resourceRest, revokeRequestRest, roleRest, roleSuiteRest)
		})
}

func initV1(
	route *echo.Group,
	authzSvc it.AuthorizeService,
	actionRest *v1.ActionRest,
	authorizedRest *v1.AuthorizeRest,
	entitlementRest *v1.EntitlementRest,
	grantRequestRest *v1.GrantRequestRest,
	resourceRest *v1.ResourceRest,
	revokeRequestRest *v1.RevokeRequestRest,
	roleRest *v1.RoleRest,
	roleSuiteRest *v1.RoleSuiteRest,
) {
	protected := route.Group("", commonMiddleware.RequireAuthMiddleware())
	checker := authzMiddleware.NewPermissionCheckerAdapter(authzSvc)

	mwActionView := commonMiddleware.RequirePermission(checker, constants.ResourceAction, constants.ActionView, nil)
	mwActionCreate := commonMiddleware.RequirePermission(checker, constants.ResourceAction, constants.ActionCreate, nil)
	mwActionUpdate := commonMiddleware.RequirePermission(checker, constants.ResourceAction, constants.ActionUpdate, nil)
	mwActionDelete := commonMiddleware.RequirePermission(checker, constants.ResourceAction, constants.ActionDelete, nil)

	mwEntitlementView := commonMiddleware.RequirePermission(checker, constants.ResourceEntitlement, constants.ActionView, nil)
	mwEntitlementCreate := commonMiddleware.RequirePermission(checker, constants.ResourceEntitlement, constants.ActionCreate, nil)
	mwEntitlementUpdate := commonMiddleware.RequirePermission(checker, constants.ResourceEntitlement, constants.ActionUpdate, nil)
	mwEntitlementDelete := commonMiddleware.RequirePermission(checker, constants.ResourceEntitlement, constants.ActionDelete, nil)

	mwResourceView := commonMiddleware.RequirePermission(checker, constants.ResourceResource, constants.ActionView, nil)
	mwResourceCreate := commonMiddleware.RequirePermission(checker, constants.ResourceResource, constants.ActionCreate, nil)
	mwResourceUpdate := commonMiddleware.RequirePermission(checker, constants.ResourceResource, constants.ActionUpdate, nil)
	mwResourceDelete := commonMiddleware.RequirePermission(checker, constants.ResourceResource, constants.ActionDelete, nil)

	mwRoleView := commonMiddleware.RequirePermission(checker, constants.ResourceRole, constants.ActionView, nil)
	mwRoleCreate := commonMiddleware.RequirePermission(checker, constants.ResourceRole, constants.ActionCreate, nil)
	mwRoleUpdate := commonMiddleware.RequirePermission(checker, constants.ResourceRole, constants.ActionUpdate, nil)
	mwRoleDelete := commonMiddleware.RequirePermission(checker, constants.ResourceRole, constants.ActionDelete, nil)
	mwRoleAddEntitlement := commonMiddleware.RequirePermission(checker, constants.ResourceRole, constants.ActionAddEntitlement, nil)
	mwRoleRemoveEntitlement := commonMiddleware.RequirePermission(checker, constants.ResourceRole, constants.ActionRemoveEntitlement, nil)

	mwRoleSuiteView := commonMiddleware.RequirePermission(checker, constants.ResourceRoleSuite, constants.ActionView, nil)
	mwRoleSuiteCreate := commonMiddleware.RequirePermission(checker, constants.ResourceRoleSuite, constants.ActionCreate, nil)
	mwRoleSuiteUpdate := commonMiddleware.RequirePermission(checker, constants.ResourceRoleSuite, constants.ActionUpdate, nil)
	mwRoleSuiteDelete := commonMiddleware.RequirePermission(checker, constants.ResourceRoleSuite, constants.ActionDelete, nil)

	mwGrantRequestView := commonMiddleware.RequirePermission(checker, constants.ResourceGrantRequest, constants.ActionView, nil)
	mwGrantRequestCreate := commonMiddleware.RequirePermission(checker, constants.ResourceGrantRequest, constants.ActionCreate, nil)
	mwGrantRequestDelete := commonMiddleware.RequirePermission(checker, constants.ResourceGrantRequest, constants.ActionDelete, nil)
	mwGrantRequestRespond := commonMiddleware.RequirePermission(checker, constants.ResourceGrantRequest, constants.ActionRespondGrantRequest, nil)

	mwRevokeRequestView := commonMiddleware.RequirePermission(checker, constants.ResourceRevokeRequest, constants.ActionView, nil)
	mwRevokeRequestCreate := commonMiddleware.RequirePermission(checker, constants.ResourceRevokeRequest, constants.ActionCreate, nil)
	mwRevokeRequestDelete := commonMiddleware.RequirePermission(checker, constants.ResourceRevokeRequest, constants.ActionDelete, nil)

	protected.POST("/is-authorized", authorizedRest.IsAuthorized)

	protected.POST("/actions", actionRest.CreateAction, mwActionCreate)
	protected.PUT("/actions/:id", actionRest.UpdateAction, mwActionUpdate)
	protected.GET("/actions/:id", actionRest.GetActionById, mwActionView)
	protected.GET("/actions", actionRest.SearchActions, mwActionView)
	protected.DELETE("/actions/:id", actionRest.DeleteActionHard, mwActionDelete)

	protected.POST("/entitlements", entitlementRest.CreateEntitlement, mwEntitlementCreate)
	protected.PUT("/entitlements/:id", entitlementRest.UpdateEntitlement, mwEntitlementUpdate)
	protected.GET("/entitlements/:id", entitlementRest.GetEntitlementById, mwEntitlementView)
	protected.POST("/entitlements/ids", entitlementRest.GetAllEntitlementByIds, mwEntitlementView)
	protected.GET("/entitlements", entitlementRest.SearchEntitlements, mwEntitlementView)
	protected.DELETE("/entitlements/:id", entitlementRest.DeleteEntitlementHard, mwEntitlementDelete)

	protected.POST("/resources", resourceRest.CreateResource, mwResourceCreate)
	protected.PUT("/resources/:id", resourceRest.UpdateResource, mwResourceUpdate)
	protected.GET("/resources/:name", resourceRest.GetResourceByName, mwResourceView)
	protected.GET("/resources", resourceRest.SearchResources, mwResourceView)
	protected.DELETE("/resources/:name", resourceRest.DeleteResourceHard, mwResourceDelete)

	protected.POST("/revoke-requests", revokeRequestRest.Create, mwRevokeRequestCreate)
	protected.POST("/revoke-requests/bulk", revokeRequestRest.CreateBulk, mwRevokeRequestCreate)
	protected.GET("/revoke-requests/:id", revokeRequestRest.GetById, mwRevokeRequestView)
	protected.GET("/revoke-requests", revokeRequestRest.Search, mwRevokeRequestView)
	protected.DELETE("/revoke-requests/:id", revokeRequestRest.Delete, mwRevokeRequestDelete)

	protected.POST("/roles", roleRest.CreateRole, mwRoleCreate)
	protected.PUT("/roles/:id", roleRest.UpdateRole, mwRoleUpdate)
	protected.DELETE("/roles/:id", roleRest.DeleteRoleHard, mwRoleDelete)
	protected.GET("/roles/:id", roleRest.GetRoleById, mwRoleView)
	protected.GET("/roles", roleRest.SearchRoles, mwRoleView)
	protected.POST("/roles/:id/entitlement-assignment", roleRest.AddEntitlements, mwRoleAddEntitlement)
	protected.DELETE("/roles/:id/entitlement-assignment", roleRest.RemoveEntitlements, mwRoleRemoveEntitlement)

	protected.POST("/role-suites", roleSuiteRest.CreateRoleSuite, mwRoleSuiteCreate)
	protected.PUT("/role-suites/:id", roleSuiteRest.UpdateRoleSuite, mwRoleSuiteUpdate)
	protected.DELETE("/role-suites/:id", roleSuiteRest.DeleteRoleSuite, mwRoleSuiteDelete)
	protected.GET("/role-suites/:id", roleSuiteRest.GetRoleSuiteById, mwRoleSuiteView)
	protected.GET("/role-suites", roleSuiteRest.SearchRoleSuites, mwRoleSuiteView)

	protected.POST("/grant-requests", grantRequestRest.CreateGrantRequest, mwGrantRequestCreate)
	protected.POST("/grant-requests/:id/cancel", grantRequestRest.CancelGrantRequest, mwGrantRequestView)
	protected.DELETE("/grant-requests/:id", grantRequestRest.DeleteGrantRequest, mwGrantRequestDelete)
	protected.GET("/grant-requests/:id", grantRequestRest.GetGrantRequestById, mwGrantRequestView)
	protected.GET("/grant-requests", grantRequestRest.SearchGrantRequests, mwGrantRequestView)
	protected.POST("/grant-requests/:id/respond", grantRequestRest.RespondToGrantRequest, mwGrantRequestRespond)
}
