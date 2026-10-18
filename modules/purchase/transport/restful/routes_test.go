package restful

import (
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/purchase/dynamicengines"
	v1 "github.com/sky-as-code/nikki-erp/modules/purchase/transport/restful/v1"
)

// The container names in the REST structs' tags are literals; they must be what the onions are
// registered under, or the handlers resolve nothing at boot.
func TestEngineNamesMatchTheRegisteredOnions(t *testing.T) {
	engineNames := map[string]string{
		models.ConfigurationSchemaName:      v1.ConfigurationEngineName,
		models.VendorProductPriceSchemaName: v1.VendorProductPriceEngineName,
		models.SourcingGroupSchemaName:      v1.SourcingGroupEngineName,
		models.AgreementSchemaName:          v1.AgreementEngineName,
		models.AgreementLineSchemaName:      v1.AgreementLineEngineName,
		models.PurchaseOrderSchemaName:      v1.PurchaseOrderEngineName,
		models.PurchaseOrderLineSchemaName:  v1.PurchaseOrderLineEngineName,
		models.AuditEventSchemaName:         v1.AuditEventEngineName,
	}
	for schemaName, engineName := range engineNames {
		assert.Equal(t, composable.EngineDependencyName(schemaName), engineName)
	}
	assert.Len(t, dynamicengines.SchemaNames(), len(engineNames))
}

func registeredRoutes(t *testing.T, restEngine composable.RestEngine) []string {
	t.Helper()
	echoApp := echo.New()
	require.NoError(t, restEngine.RegisterRoutes(echoApp.Group(routeV1Prefix), passThrough))

	routes := echoApp.Router().Routes()
	result := make([]string, 0, len(routes))
	for _, route := range routes {
		result = append(result, route.Method+" "+route.Path)
	}
	return result
}

func passThrough(next echo.HandlerFunc) echo.HandlerFunc {
	return next
}

// legacyBuiltinRoutes is the verb set the legacy engine served, which the clients still call.
func legacyBuiltinRoutes(schemaName string) []string {
	base := routeV1Prefix + "/" + schemaName
	return []string{
		"GET " + base + "/meta/schema",
		"POST " + base + "/exists",
		"POST " + base,
		"GET " + base,
		"POST " + base + "/meta/compute/:field",
		"POST " + base + "/:id/archived",
		"DELETE " + base + "/:id",
		"GET " + base + "/:id",
		"PATCH " + base + "/:id",
	}
}

func withActions(schemaName string, paths ...string) []string {
	routes := legacyBuiltinRoutes(schemaName)
	for _, path := range paths {
		routes = append(routes, "POST "+routeV1Prefix+"/"+schemaName+"/"+path)
	}
	return routes
}

func TestPlainResourcesKeepTheLegacyVerbSet(t *testing.T) {
	for schemaName, handlers := range map[string]composable.CrudRestHandlers{
		models.ConfigurationSchemaName:      &v1.ConfigurationRest{},
		models.VendorProductPriceSchemaName: &v1.VendorProductPriceRest{},
		models.SourcingGroupSchemaName:      &v1.SourcingGroupRest{},
		models.AgreementLineSchemaName:      &v1.AgreementLineRest{},
		models.PurchaseOrderLineSchemaName:  &v1.PurchaseOrderLineRest{},
		models.AuditEventSchemaName:         &v1.AuditEventRest{},
	} {
		t.Run(schemaName, func(t *testing.T) {
			assert.ElementsMatch(t, legacyBuiltinRoutes(schemaName), registeredRoutes(t, crudRoutes(schemaName, handlers)))
		})
	}
}

// The lifecycle actions keep the paths the legacy engine actions used.
func TestPurchaseOrderActionsKeepTheirLegacyPaths(t *testing.T) {
	expected := withActions(models.PurchaseOrderSchemaName,
		"merge", ":id/confirm", ":id/approve", ":id/cancel", ":id/send", ":id/lock", ":id/unlock",
		":id/acknowledge", ":id/create_alternative", ":id/compare_alternatives", ":id/reprice",
		":id/duplicate")
	assert.ElementsMatch(t, expected, registeredRoutes(t, purchaseOrderRoutes(&v1.PurchaseOrderRest{})))
}

func TestAgreementActionsKeepTheirLegacyPaths(t *testing.T) {
	expected := withActions(models.AgreementSchemaName,
		":id/confirm", ":id/close", ":id/cancel", ":id/create_rfq")
	assert.ElementsMatch(t, expected, registeredRoutes(t, agreementRoutes(&v1.AgreementRest{})))
}
