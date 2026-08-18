package restful

import (
	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/contacts/dynamicengines"
	m "github.com/sky-as-code/nikki-erp/modules/core/httpserver/middlewares"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

func InitRestfulHandlers() error {
	return initContactsV1()
}

// initContactsV1 mounts the engine routes under /v1/contacts.
//
// This replaced a hand-written route table under "/v1/:org_id/contacts", where a party's channels
// and relationships were nested paths ("/parties/:party_id/channels"). The engine addresses every
// resource by its own schema name instead, so those nested lists became graph filters on party_id
// and the organization moved from the path into a query parameter.
func initContactsV1() error {
	return deps.Invoke(func(route *echo.Group) error {
		registerEngineRoutes(route.Group("/v1/contacts"))
		return nil
	})
}

// registerEngineRoutes exposes every Contacts resource engine over HTTP.
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
