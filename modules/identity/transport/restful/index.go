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
	deps.Register(v1.NewUserRest)
	return deps.Invoke(func(route *echo.Group, userRest *v1.UserRest) {
		v1 := route.Group("/v1/identity")
		initV1(v1, userRest)
	})
}

func initV1(route *echo.Group, userRest *v1.UserRest) {
	route.POST("/users", userRest.CreateUser)
	route.PUT("/users/:id", userRest.UpdateUser)
}
