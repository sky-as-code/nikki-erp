package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	modconstants "github.com/sky-as-code/nikki-erp/modules/accounting/constants"
	"github.com/sky-as-code/nikki-erp/modules/accounting/dynamicengines"
	v1 "github.com/sky-as-code/nikki-erp/modules/accounting/transport/restful/v1"
	m "github.com/sky-as-code/nikki-erp/modules/core/httpserver/middlewares"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

func InitRestfulHandlers() error {
	return stdErr.Join(
		deps.Register(v1.NewTaxCalculationRest),
		initAccountingV1(),
	)
}

func initAccountingV1() error {
	return deps.Invoke(func(route *echo.Group) error {
		routeV1 := route.Group(modconstants.AccountingRouteV1)
		registerEngineRoutes(routeV1)
		return initTaxCalculationV1(routeV1)
	})
}

// registerEngineRoutes exposes every Accounting resource engine over HTTP.
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

// initTaxCalculationV1 registers the endpoints the resource engines cannot express. They hang off
// "/tax/" rather than the "accounting_tax" resource path because they are not operations on a tax
// record, and would collide with the engine's own ":id" route.
func initTaxCalculationV1(routeV1 *echo.Group) error {
	return deps.Invoke(func(taxRest *v1.TaxCalculationRest) error {
		// SmokeAuthz loads the caller's identity and entitlements into the request context. Without
		// it every endpoint here sees an empty permission set and denies the request, including an
		// owner's, whose bypass depends on a flag this middleware sets.
		//
		// POST throughout, including for the two that only read: the request is a whole document of
		// lines and party context, which does not fit a query string.
		routeV1.POST("/tax/calculate", taxRest.Calculate, m.SmokeAuthz())
		routeV1.POST("/tax/simulate", taxRest.Simulate, m.SmokeAuthz())
		routeV1.POST("/tax/reverse-full", taxRest.ReverseFull, m.SmokeAuthz())
		routeV1.POST("/tax/reverse-partial", taxRest.ReversePartial, m.SmokeAuthz())
		return nil
	})
}
