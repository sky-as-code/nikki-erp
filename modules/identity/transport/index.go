package transport

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/core/config"
	// c "github.com/sky-as-code/nikki-erp/modules/core/constants"
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	v1 "github.com/sky-as-code/nikki-erp/modules/identity/transport/restful/v1"
)

type transportParams struct {
	dig.In

	Config    config.ConfigService
	Logger    logging.LoggerService
	RootRoute *echo.Group
}

func InitTransport(params transportParams) error {
	deps.Register(v1.NewUserRest)
	return deps.Invoke(func(userRest *v1.UserRest) {
		route := params.RootRoute
		v1 := route.Group("/v1")
		initV1(v1, userRest)
	})
}

func initV1(route *echo.Group, userRest *v1.UserRest) {
	route.POST("/users", userRest.CreateUser)
}
