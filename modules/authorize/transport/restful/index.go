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
	deps.Register(v1.NewResourceRest, v1.NewActionRest, v1.NewEntitlementRest, v1.NewRoleRest, v1.NewRoleSuiteRest, v1.NewAuthorizeRest)
	return deps.Invoke(func(route *echo.Group, resourceRest *v1.ResourceRest, actionRest *v1.ActionRest, entitlementRest *v1.EntitlementRest, roleRest *v1.RoleRest, roleSuiteRest *v1.RoleSuiteRest, authorizedRest *v1.AuthorizeRest) {
		v1 := route.Group("/v1/authorize")
		initV1(v1, resourceRest, actionRest, entitlementRest, roleRest, roleSuiteRest, authorizedRest)
	})
}

func initV1(route *echo.Group, resourceRest *v1.ResourceRest, actionRest *v1.ActionRest, entitlementRest *v1.EntitlementRest, roleRest *v1.RoleRest, roleSuiteRest *v1.RoleSuiteRest, authorizedRest *v1.AuthorizeRest) {
	route.POST("/resources", resourceRest.CreateResource)
	route.PUT("/resources/:id", resourceRest.UpdateResource)
	route.DELETE("/resources/:id", resourceRest.DeleteHardResource)
	route.GET("/resources/:name", resourceRest.GetResourceByName)
	route.GET("/resources", resourceRest.SearchResources)

	route.POST("/actions", actionRest.CreateAction)
	route.PUT("/actions/:id", actionRest.UpdateAction)
	route.DELETE("/actions/:id", actionRest.DeleteHardAction)
	route.GET("/actions/:id", actionRest.GetActionById)
	route.GET("/actions", actionRest.SearchActions)

	route.POST("/entitlements", entitlementRest.CreateEntitlement)
	route.PUT("/entitlements/:id", entitlementRest.UpdateEntitlement)
	route.DELETE("/entitlements/:id", entitlementRest.DeleteHardEntitlement)
	route.GET("/entitlements/:id", entitlementRest.GetEntitlementById)
	route.GET("/entitlements", entitlementRest.SearchEntitlements)

	route.POST("/roles", roleRest.CreateRole)
	route.GET("/roles/:id", roleRest.GetRoleById)
	route.GET("/roles", roleRest.SearchRoles)

	route.POST("/role-suites", roleSuiteRest.CreateRoleSuite)
	route.GET("/role-suites/:id", roleSuiteRest.GetRoleSuiteById)
	route.GET("/role-suites", roleSuiteRest.SearchRoleSuites)

	route.POST("/isauthorized", authorizedRest.IsAuthorized)
}
