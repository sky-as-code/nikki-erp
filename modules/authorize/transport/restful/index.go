package restful

import (
	"errors"

	"github.com/labstack/echo/v4"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
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
	route.POST("/actions", actionRest.CreateAction)
	route.PUT("/actions/:id", actionRest.UpdateAction)
	route.GET("/actions/:id", actionRest.GetActionById)
	route.GET("/actions", actionRest.SearchActions)
	route.DELETE("/actions/:id", actionRest.DeleteActionHard)

	route.POST("/isauthorized", authorizedRest.IsAuthorized)

	route.POST("/entitlements", entitlementRest.CreateEntitlement)
	route.PUT("/entitlements/:id", entitlementRest.UpdateEntitlement)
	route.GET("/entitlements/:id", entitlementRest.GetEntitlementById)
	route.POST("/entitlements/ids", entitlementRest.GetAllEntitlementByIds)
	route.GET("/entitlements", entitlementRest.SearchEntitlements)
	route.DELETE("/entitlements/:id", entitlementRest.DeleteEntitlementHard)

	route.POST("/grant-requests", grantRequestRest.CreateGrantRequest)
	route.POST("/grant-requests/:id/cancel", grantRequestRest.CancelGrantRequest)
	// delete api
	route.POST("/grant-requests/:id/respond", grantRequestRest.RespondToGrantRequest)

	route.POST("/resources", resourceRest.CreateResource)
	route.PUT("/resources/:id", resourceRest.UpdateResource)
	route.GET("/resources/:name", resourceRest.GetResourceByName)
	route.GET("/resources", resourceRest.SearchResources)
	route.DELETE("/resources/:name", resourceRest.DeleteResourceHard)

	route.POST("/revoke-requests", revokeRequestRest.Create)
	route.POST("/revoke-requests/bulk", revokeRequestRest.CreateBulk)
	route.GET("/revoke-requests/:id", revokeRequestRest.GetById)
	route.GET("/revoke-requests", revokeRequestRest.Search)
	route.DELETE("/revoke-requests/:id", revokeRequestRest.Delete)

	route.POST("/roles", roleRest.CreateRole)
	route.PUT("/roles/:id", roleRest.UpdateRole)
	route.DELETE("/roles/:id", roleRest.DeleteRoleHard)
	route.GET("/roles/:id", roleRest.GetRoleById)
	route.GET("/roles", roleRest.SearchRoles)
	route.POST("/roles/:id/entitlement-assignment", roleRest.AddEntitlements)
	route.DELETE("/roles/:id/entitlement-assignment", roleRest.RemoveEntitlements)

	route.POST("/role-suites", roleSuiteRest.CreateRoleSuite)
	route.PUT("/role-suites/:id", roleSuiteRest.UpdateRoleSuite)
	route.DELETE("/role-suites/:id", roleSuiteRest.DeleteRoleSuite)
	route.GET("/role-suites/:id", roleSuiteRest.GetRoleSuiteById)
	route.GET("/role-suites", roleSuiteRest.SearchRoleSuites)

	route.POST("/isauthorized", authorizedRest.IsAuthorized)

	route.POST("/grant-requests", grantRequestRest.CreateGrantRequest)
	route.POST("/grant-requests/:id/cancel", grantRequestRest.CancelGrantRequest)
	route.DELETE("/grant-requests/:id", grantRequestRest.DeleteGrantRequest)
	route.GET("/grant-requests/:id", grantRequestRest.GetGrantRequestById)
	route.GET("/grant-requests", grantRequestRest.SearchGrantRequests)
	route.POST("/grant-requests/:id/respond", grantRequestRest.RespondToGrantRequest)
}
