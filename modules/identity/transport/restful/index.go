package restful

import (
	"github.com/labstack/echo/v4"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	v1 "github.com/sky-as-code/nikki-erp/modules/identity/transport/restful/v1"
)

func InitRestfulHandlers() error {
	deps.Register(
		v1.NewUserRest,
		v1.NewGroupRest,
		v1.NewOrganizationRest,
		v1.NewHierarchyRest,
	)
	return deps.Invoke(func(
		route *echo.Group,
		cqrsBus cqrs.CqrsBus,
		userRest *v1.UserRest,
		groupRest *v1.GroupRest,
		orgRest *v1.OrganizationRest,
		hierarchyRest *v1.HierarchyRest,
	) {
		v1 := route.Group("/v1/identity")
		initV1(v1, cqrsBus, userRest, groupRest, orgRest, hierarchyRest)
	})
}

func initV1(
	route *echo.Group,
	cqrsBus cqrs.CqrsBus,
	userRest *v1.UserRest,
	groupRest *v1.GroupRest,
	orgRest *v1.OrganizationRest,
	hierarchyRest *v1.HierarchyRest,
) {
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

	// mwHierarchyView := commonMiddleware.RequirePermission(checker, constants.ResourceHierarchyLevel, constants.ActionView, nil)
	// mwHierarchyCreate := commonMiddleware.RequirePermission(checker, constants.ResourceHierarchyLevel, constants.ActionCreate, nil)
	// mwHierarchyUpdate := commonMiddleware.RequirePermission(checker, constants.ResourceHierarchyLevel, constants.ActionUpdate, nil)
	// mwHierarchyDelete := commonMiddleware.RequirePermission(checker, constants.ResourceHierarchyLevel, constants.ActionDelete, nil)
	// mwHierarchyManageUsers := commonMiddleware.RequirePermission(checker, constants.ResourceHierarchyLevel, constants.ActionManageUsers, nil)

	// protected.POST("/users", userRest.CreateUser, mwUserCreate)
	// protected.DELETE("/users/:id", userRest.DeleteUser, mwUserDelete)
	// protected.GET("/users/:id", userRest.GetUserById, mwUserView)
	// protected.GET("/users", userRest.SearchUsers, mwUserView)
	// protected.PUT("/users/:id", userRest.UpdateUser, mwUserUpdate)
	// protected.POST("/users/exists", userRest.UserExistsMulti, mwUserView)

	// protected.POST("/groups", groupRest.CreateGroup, mwGroupCreate)
	// protected.DELETE("/groups/:id", groupRest.DeleteGroup, mwGroupDelete)
	// protected.GET("/groups/:id", groupRest.GetGroupById, mwGroupView)
	// protected.GET("/groups", groupRest.SearchGroups, mwGroupView)
	// protected.PUT("/groups/:id", groupRest.UpdateGroup, mwGroupUpdate)
	// protected.POST("/groups/:groupId/manage-users", groupRest.ManageGroupUsers, mwGroupUpdate)

	// protected.POST("/organizations", orgRest.CreateOrganization, mwOrgCreate)
	// protected.DELETE("/organizations/:slug", orgRest.DeleteOrganization, mwOrgDelete)
	// protected.GET("/organizations/:slug", orgRest.GetOrganizationBySlug, mwOrgView)
	// protected.GET("/organizations", orgRest.SearchOrganizations, mwOrgView)
	// protected.PUT("/organizations/:slug", orgRest.UpdateOrganization, mwOrgUpdate)
	// protected.POST("/organizations/:orgId/manage-users", orgRest.ManageOrganizationUsers, mwOrgUpdate)

	// protected.POST("/hierarchy", hierarchyRest.CreateHierarchyLevel, mwHierarchyCreate)
	// protected.DELETE("/hierarchy/:id", hierarchyRest.DeleteHierarchyLevel, mwHierarchyDelete)
	// protected.GET("/hierarchy/:id", hierarchyRest.GetHierarchyLevelById, mwHierarchyView)
	// protected.GET("/hierarchy", hierarchyRest.SearchHierarchyLevels, mwHierarchyView)
	// protected.PUT("/hierarchy/:id", hierarchyRest.UpdateHierarchyLevel, mwHierarchyUpdate)
	// protected.POST("/hierarchy/:hierarchyId/manage-users", hierarchyRest.ManageHierarchyUsers, mwHierarchyManageUsers)

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
	route.POST("/organizations/:orgId/manage-users", orgRest.ManageOrganizationUsers)

	route.POST("/hierarchy", hierarchyRest.CreateHierarchyLevel)
	route.DELETE("/hierarchy/:id", hierarchyRest.DeleteHierarchyLevel)
	route.GET("/hierarchy/:id", hierarchyRest.GetHierarchyLevelById)
	route.GET("/hierarchy", hierarchyRest.SearchHierarchyLevels)
	route.PUT("/hierarchy/:id", hierarchyRest.UpdateHierarchyLevel)
	route.POST("/hierarchy/:hierarchyId/manage-users", hierarchyRest.ManageHierarchyUsers)

}
