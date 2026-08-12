package restful

import (
	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	m "github.com/sky-as-code/nikki-erp/modules/core/httpserver/middlewares"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/inventory/dynamicengines"
)

func InitRestfulHandlers() error {
	return initInventoryV1()
}

func initInventoryV1() error {
	return deps.Invoke(func(
		route *echo.Group,
	) error {
		routeV1 := route.Group("/v1/inventory")

		// Inventory has no hand-written REST layer at all. Every route — CRUD and the Products
		// capabilities alike — is declared on an engine as an action and registered here, so
		// that the permission check, validation pipeline and response envelope are the engine's
		// for all of them. See dynamicengines/product_template.go and product_variant.go.
		registerEngineRoutes(routeV1)

		return nil
	})
}

// registerEngineRoutes exposes every Inventory resource engine over HTTP.
// A missing engine is skipped, so that a build which drops one still starts.
func registerEngineRoutes(routeV1 *echo.Group) {
	for _, schemaName := range dynamicengines.EngineSchemaNames() {
		engine, exists := dynamicresource.Registry().GetEngine(schemaName)
		if !exists {
			continue
		}
		engine.RestApi().RegisterRoutes(routeV1, m.SmokeAuthz())
	}
}
