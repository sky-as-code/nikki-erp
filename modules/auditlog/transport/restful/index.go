package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	c "github.com/sky-as-code/nikki-erp/modules/auditlog/constants"
	"github.com/sky-as-code/nikki-erp/modules/auditlog/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/auditlog/dynamicengines"
	v1 "github.com/sky-as-code/nikki-erp/modules/auditlog/transport/restful/v1"
	m "github.com/sky-as-code/nikki-erp/modules/core/httpserver/middlewares"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

func InitRestfulHandlers() error {
	return stdErr.Join(
		deps.Register(v1.NewActionRest),
		initAuditLogV1(),
	)
}

func initAuditLogV1() error {
	return deps.Invoke(func(route *echo.Group, actionRest *v1.ActionRest) error {
		routeV1 := route.Group("/v1/" + c.AuditLogModuleName)
		routeV1.GET(c.RouteActionHistory, actionRest.GetHistory, m.UserSmokeAuthz())
		return composable.NewRestEngine(models.ActionSchemaName, actionRest).
			AddCrudRoutes(dynamicengines.ActionRestActions()...).
			RegisterRoutes(routeV1)
	})
}
