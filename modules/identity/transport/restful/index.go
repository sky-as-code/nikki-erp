package restful

import (
	"github.com/labstack/echo/v4"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/middleware"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
	v1 "github.com/sky-as-code/nikki-erp/modules/identity/transport/restful/v1"
)

func InitRestfulHandlers() error {
	deps.Register(
		v1.NewUserRest,
		v1.NewGroupRest,
		v1.NewOrganizationRest,
		v1.NewHierarchyRest,
	)
	return deps.Invoke(func(route *echo.Group, userRest *v1.UserRest, groupRest *v1.GroupRest, orgRest *v1.OrganizationRest, hierarchyRest *v1.HierarchyRest) {
		v1 := route.Group("/v1/identity")
		initV1(v1, userRest, groupRest, orgRest, hierarchyRest)
	})
}

func initV1(route *echo.Group, userRest *v1.UserRest, groupRest *v1.GroupRest, orgRest *v1.OrganizationRest, hierarchyRest *v1.HierarchyRest) {
	route.POST(
		"/users/dynamic",
		middleware.DynamicAutoMapper[it.CreateUserCommand](userRest.DynamicCreateUser),
		middleware.DynamicValidator("identity.createUserRequest"),
	)

	route.POST("/users", userRest.CreateUser)
	route.DELETE("/users/:id", userRest.DeleteUser)
	route.GET("/users/:id", userRest.GetUserById)
	route.GET("/users", userRest.SearchUsers)
	route.PUT("/users/:id", userRest.UpdateUser)
	route.POST("/users/exists", userRest.UserExistsMulti)

	route.POST("/groups", groupRest.CreateGroup)
	route.DELETE("/groups/:id", groupRest.DeleteGroup)
	route.GET("/groups/:id", groupRest.GetGroupById)
	route.GET("/groups", groupRest.SearchGroups)
	route.PUT("/groups/:id", groupRest.UpdateGroup)
	route.POST("/groups/:groupId/manage-users", groupRest.ManageGroupUsers)

	route.POST("/organizations", orgRest.CreateOrganization)
	route.DELETE("/organizations/:slug", orgRest.DeleteOrganization)
	route.GET("/organizations/:slug", orgRest.GetOrganizationBySlug)
	route.GET("/organizations", orgRest.SearchOrganizations)
	route.PUT("/organizations/:slug", orgRest.UpdateOrganization)

	route.POST("/hierarchy", hierarchyRest.CreateHierarchyLevel)
	route.DELETE("/hierarchy/:id", hierarchyRest.DeleteHierarchyLevel)
	route.GET("/hierarchy/:id", hierarchyRest.GetHierarchyLevelById)
	route.GET("/hierarchy", hierarchyRest.SearchHierarchyLevels)
	route.PUT("/hierarchy/:id", hierarchyRest.UpdateHierarchyLevel)
	route.POST("/hierarchy/:hierarchyId/manage-users", hierarchyRest.ManageHierarchyUsers)
}
