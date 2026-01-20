package restful

import (
	"errors"

	"github.com/labstack/echo/v4"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	commonMiddleware "github.com/sky-as-code/nikki-erp/common/middleware"
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
			initV1(v1, actionRest, authorizedRest, entitlementRest, grantRequestRest, resourceRest, revokeRequestRest, roleRest, roleSuiteRest)
		})
}

func initV1(
	route *echo.Group,
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

	protected.POST("/isauthorized", authorizedRest.IsAuthorized)

	protected.POST("/actions", actionRest.CreateAction)
	protected.PUT("/actions/:id", actionRest.UpdateAction)
	protected.GET("/actions/:id", actionRest.GetActionById)
	protected.GET("/actions", actionRest.SearchActions)
	protected.DELETE("/actions/:id", actionRest.DeleteActionHard)

	protected.POST("/entitlements", entitlementRest.CreateEntitlement)
	protected.PUT("/entitlements/:id", entitlementRest.UpdateEntitlement)
	protected.GET("/entitlements/:id", entitlementRest.GetEntitlementById)
	protected.POST("/entitlements/ids", entitlementRest.GetAllEntitlementByIds)
	protected.GET("/entitlements", entitlementRest.SearchEntitlements)
	protected.DELETE("/entitlements/:id", entitlementRest.DeleteEntitlementHard)

	protected.POST("/resources", resourceRest.CreateResource)
	protected.PUT("/resources/:id", resourceRest.UpdateResource)
	protected.GET("/resources/:name", resourceRest.GetResourceByName)
	protected.GET("/resources", resourceRest.SearchResources)
	protected.DELETE("/resources/:name", resourceRest.DeleteResourceHard)

	protected.POST("/revoke-requests", revokeRequestRest.Create)
	protected.POST("/revoke-requests/bulk", revokeRequestRest.CreateBulk)
	protected.GET("/revoke-requests/:id", revokeRequestRest.GetById)
	protected.GET("/revoke-requests", revokeRequestRest.Search)
	protected.DELETE("/revoke-requests/:id", revokeRequestRest.Delete)

	protected.POST("/roles", roleRest.CreateRole)
	protected.PUT("/roles/:id", roleRest.UpdateRole)
	protected.DELETE("/roles/:id", roleRest.DeleteRoleHard)
	protected.GET("/roles/:id", roleRest.GetRoleById)
	protected.GET("/roles", roleRest.SearchRoles)
	protected.POST("/roles/:id/entitlement-assignment", roleRest.AddEntitlements)
	protected.DELETE("/roles/:id/entitlement-assignment", roleRest.RemoveEntitlements)

	protected.POST("/role-suites", roleSuiteRest.CreateRoleSuite)
	protected.PUT("/role-suites/:id", roleSuiteRest.UpdateRoleSuite)
	protected.DELETE("/role-suites/:id", roleSuiteRest.DeleteRoleSuite)
	protected.GET("/role-suites/:id", roleSuiteRest.GetRoleSuiteById)
	protected.GET("/role-suites", roleSuiteRest.SearchRoleSuites)

	protected.POST("/grant-requests", grantRequestRest.CreateGrantRequest)
	protected.POST("/grant-requests/:id/cancel", grantRequestRest.CancelGrantRequest)
	protected.DELETE("/grant-requests/:id", grantRequestRest.DeleteGrantRequest)
	protected.GET("/grant-requests/:id", grantRequestRest.GetGrantRequestById)
	protected.GET("/grant-requests", grantRequestRest.SearchGrantRequests)
	protected.POST("/grant-requests/:id/respond", grantRequestRest.RespondToGrantRequest)
}
