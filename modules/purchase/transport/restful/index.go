package restful

import (
	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	m "github.com/sky-as-code/nikki-erp/modules/core/httpserver/middlewares"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/purchase/dynamicengines"
)

func InitRestfulHandlers() error {
	return initPurchaseV1()
}

func initPurchaseV1() error {
	return deps.Invoke(func(route *echo.Group) error {
		registerEngineRoutes(route.Group("/v1/purchase"))
		return nil
	})
}

// A missing engine is skipped, so a build that drops one still starts.
func registerEngineRoutes(routeV1 *echo.Group) {
	for _, schemaName := range dynamicengines.EngineSchemaNames() {
		engine, exists := dynamicresource.Registry().GetEngine(schemaName)
		if !exists {
			continue
		}
		engine.RestApi().RegisterRoutes(routeV1, m.SmokeAuthz())
	}
}
