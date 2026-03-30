package restful

import (
	"github.com/labstack/echo/v4"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver/middlewares"
	"github.com/sky-as-code/nikki-erp/modules/identity/constants"
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
		v1.Use(middlewares.RequestContextMiddleware2(constants.IdentityModuleName))
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

	route.DELETE("/users/:id", userRest.DeleteUser)
	route.GET("/users/:id", userRest.GetUser)
	route.GET("/users", userRest.SearchUsers)
	route.POST("/users/exists", userRest.UserExists)
	route.POST("/users/:id/archived", userRest.SetUserIsArchived)
	route.POST("/users", userRest.CreateUser)
	route.PUT("/users/:id", userRest.UpdateUser)
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

	route.DELETE("/groups/:id", groupRest.DeleteGroup)
	route.GET("/groups/:id", groupRest.GetGroup)
	route.GET("/groups", groupRest.SearchGroups)
	route.POST("/groups/exists", groupRest.GroupExists)
	route.POST("/groups/:group_id/manage-users", groupRest.ManageGroupUsers)
	route.POST("/groups", groupRest.CreateGroup)
	route.PUT("/groups/:id", groupRest.UpdateGroup)

	route.DELETE("/organizations/:id", orgRest.DeleteOrg)
	route.GET("/organizations/:id", orgRest.GetOrg)
	route.GET("/organizations", orgRest.SearchOrgs)
	route.POST("/organizations/:id/archived", orgRest.SetOrgIsArchived)
	route.POST("/organizations/:org_id/manage-users", orgRest.ManageOrgUsers)
	route.POST("/organizations/exists", orgRest.OrgExists)
	route.POST("/organizations", orgRest.CreateOrg)
	route.PUT("/organizations/:id", orgRest.UpdateOrg)

	route.DELETE("/hierarchies/:id", hierarchyRest.DeleteHierarchyLevel)
	route.GET("/hierarchies/:id", hierarchyRest.GetHierarchyLevel)
	route.GET("/hierarchies", hierarchyRest.SearchHierarchyLevels)
	route.POST("/hierarchies/:id/exists", hierarchyRest.HierarchyLevelExists)
	route.POST("/hierarchies/:hierarchy_id/manage-users", hierarchyRest.ManageHierarchyUsers)
	route.POST("/hierarchies", hierarchyRest.CreateHierarchyLevel)
	route.PUT("/hierarchies/:id", hierarchyRest.UpdateHierarchyLevel)
}
