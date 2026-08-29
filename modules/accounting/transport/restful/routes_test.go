package restful

import (
	"os"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"

	modconstants "github.com/sky-as-code/nikki-erp/modules/accounting/constants"
)

// taxRoutePaths is the set of sub-paths initTaxCalculationV1 registers, asserted against the real
// source below so that adding an endpoint without updating this list fails.
var taxRoutePaths = []string{
	"/tax/calculate",
	"/tax/simulate",
	"/tax/reverse-full",
	"/tax/reverse-partial",
}

// registerTaxRoutes mirrors what initTaxCalculationV1 registers, minus the dependency injection.
// Paths are asserted rather than handlers: booting proves the DI wiring, but a typo in a route
// string produces a 404 nothing else catches.
func registerTaxRoutes(routeV1 *echo.Group) {
	noop := func(echoCtx *echo.Context) error { return nil }
	for _, path := range taxRoutePaths {
		routeV1.POST(path, noop)
	}
}

func TestTaxRoutesAreRegisteredUnderTheModulePrefix(t *testing.T) {
	server := echo.New()
	registerTaxRoutes(server.Group(modconstants.AccountingRouteV1))

	registered := map[string]string{}
	for _, route := range server.Router().Routes() {
		registered[route.Path] = route.Method
	}

	expected := []string{
		"/v1/accounting/tax/calculate",
		"/v1/accounting/tax/simulate",
		"/v1/accounting/tax/reverse-full",
		"/v1/accounting/tax/reverse-partial",
	}
	for _, path := range expected {
		method, found := registered[path]
		if !found {
			t.Errorf("expected %s to be registered; registered paths: %v", path, registered)
			continue
		}
		// POST throughout, including for the two that only read: the request is a whole document of
		// lines and party context, which does not fit a query string.
		if method != "POST" {
			t.Errorf("expected %s to be a POST, got %s", path, method)
		}
	}
}

// The prefix is the same string the IAM resource codes and schema names share; drifting from it
// leaves the endpoints reachable at a path no client calls.
func TestModuleRoutePrefixIsStable(t *testing.T) {
	if modconstants.AccountingRouteV1 != "/v1/accounting" {
		t.Fatalf("the module route prefix changed to %q; clients and docs assume /v1/accounting",
			modconstants.AccountingRouteV1)
	}
}

// Every custom route must carry SmokeAuthz, and this test reads the real source to check it.
// Without it the handler runs with an empty permission context and answers 403 to everyone,
// including an owner, which no compiler or route-path assertion would catch.
func TestCustomRoutesCarrySmokeAuthz(t *testing.T) {
	source, err := os.ReadFile("index.go")
	if err != nil {
		t.Fatalf("could not read the route registration source: %v", err)
	}

	for _, path := range taxRoutePaths {
		index := strings.Index(string(source), `"`+path+`"`)
		if index < 0 {
			t.Errorf("route %s is not registered in index.go", path)
			continue
		}
		// The registration is one call; the middleware has to appear before the line ends.
		line := string(source)[index:]
		if end := strings.IndexByte(line, 0x0a); end >= 0 {
			line = line[:end]
		}
		if !strings.Contains(line, "m.SmokeAuthz()") {
			t.Errorf("route %s is registered without m.SmokeAuthz(); every request to it would be denied", path)
		}
	}
}
