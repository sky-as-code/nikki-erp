package restful

import (
	"errors"

	"github.com/labstack/echo/v4"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	v1 "github.com/sky-as-code/nikki-erp/modules/identity/transport/restful/v1"
)

func InitRestfulHandlers() error {
	err := errors.Join(
		initUserRest(),
	)
	return err
}

func initUserRest() error {
	deps.Register(v1.NewUserRest, v1.NewGroupRest)
	return deps.Invoke(func(route *echo.Group, userRest *v1.UserRest, groupRest *v1.GroupRest) {
		v1 := route.Group("/v1/identity")
		initV1(v1, userRest, groupRest)
	})
}

func initV1(route *echo.Group, userRest *v1.UserRest, groupRest *v1.GroupRest) {
	route.POST("/users", userRest.CreateUser)
	route.DELETE("/users/:id", userRest.DeleteUser)
	route.GET("/users/:id", userRest.GetUserById)
	route.PUT("/users/:id", userRest.UpdateUser)

	route.POST("/groups", groupRest.CreateGroup)
	route.DELETE("/groups/:id", groupRest.DeleteGroup)
	route.GET("/groups/:id", groupRest.GetGroupById)
	route.PUT("/groups/:id", groupRest.UpdateGroup)
}
