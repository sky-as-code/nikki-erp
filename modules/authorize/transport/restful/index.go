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
	deps.Register(v1.NewResourceRest, v1.NewActionRest, v1.NewEntitlementRest, v1.NewRoleRest)
	return deps.Invoke(func(route *echo.Group, resourceRest *v1.ResourceRest, actionRest *v1.ActionRest, entitlementRest *v1.EntitlementRest, roleRest *v1.RoleRest) {
		v1 := route.Group("/v1/authorize")
		initV1(v1, resourceRest, actionRest, entitlementRest, roleRest)
	})
}

func initV1(route *echo.Group, resourceRest *v1.ResourceRest, actionRest *v1.ActionRest, entitlementRest *v1.EntitlementRest, roleRest *v1.RoleRest) {
	// START: Resource
	route.POST("/resources", resourceRest.CreateResource)
	route.PUT("/resources/:id", resourceRest.UpdateResource)
	route.GET("/resources/:name", resourceRest.GetResourceByName)
	route.GET("/resources", resourceRest.SearchResources)
	// Implement after complete all the authorization logic
	// route.DELETE("/resources/:id", resourceRest.DeleteResource)

	// END: Resource

	// START: Action
	route.POST("/actions", actionRest.CreateAction)
	route.PUT("/actions/:id", actionRest.UpdateAction)
	route.GET("/actions/:id", actionRest.GetActionById)
	route.GET("/actions", actionRest.SearchActions)
	// Implement after complete all the authorization logic
	// route.DELETE("/actions/:id", actionRest.DeleteAction)

	// END: Action

	// START: Entitlement
	route.POST("/entitlements", entitlementRest.CreateEntitlement)
	route.PUT("/entitlements/:id", entitlementRest.UpdateEntitlement)
	route.GET("/entitlements/:id", entitlementRest.GetEntitlementById)
	route.GET("/entitlements", entitlementRest.SearchEntitlements) // Name can null, that why we can't use the name fied in graph

	// END: Entitlement

	// START: Role
	route.POST("/roles", roleRest.CreateRole)
	route.GET("/roles/:id", roleRest.GetRoleById)
	route.GET("/roles", roleRest.SearchRoles)
	// END: Role
}
