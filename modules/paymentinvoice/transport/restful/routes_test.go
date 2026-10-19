package restful

import (
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/dynamicengines"
	v1 "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/transport/restful/v1"
)

// The container names in the REST structs' tags are literals; they must be what the onions are
// registered under, or the handlers resolve nothing at boot.
func TestEngineNamesMatchTheRegisteredOnions(t *testing.T) {
	for schemaName, engineName := range map[string]string{
		models.PaymentMethodSchemaName:  v1.PaymentMethodEngineName,
		models.PaymentProfileSchemaName: v1.PaymentProfileEngineName,
		models.TransactionSchemaName:    v1.TransactionEngineName,
		models.OrderSchemaName:          v1.OrderEngineName,
		models.InvoiceSchemaName:        v1.InvoiceEngineName,
		models.InvoiceLineSchemaName:    v1.InvoiceLineEngineName,
	} {
		assert.Equal(t, composable.EngineDependencyName(schemaName), engineName)
	}
	assert.Len(t, dynamicengines.SchemaNames(), 6)
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

// legacyBuiltinRoutes is the verb set the legacy engine served, which the clients still call:
// no bulk create and no import.
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

func TestPlainResourcesKeepTheLegacyVerbSet(t *testing.T) {
	for schemaName, handlers := range map[string]composable.CrudRestHandlers{
		models.PaymentMethodSchemaName:  &v1.PaymentMethodRest{},
		models.PaymentProfileSchemaName: &v1.PaymentProfileRest{},
		models.TransactionSchemaName:    &v1.TransactionRest{},
		models.InvoiceLineSchemaName:    &v1.InvoiceLineRest{},
	} {
		t.Run(schemaName, func(t *testing.T) {
			assert.ElementsMatch(t, legacyBuiltinRoutes(schemaName), registeredRoutes(t, crudRoutes(schemaName, handlers)))
		})
	}
}

// The money actions keep the paths the vending machines and the ordering system call.
func TestOrderActionsKeepTheirLegacyPaths(t *testing.T) {
	base := routeV1Prefix + "/" + models.OrderSchemaName
	expected := append(legacyBuiltinRoutes(models.OrderSchemaName),
		"POST "+base+"/create_payment",
		"POST "+base+"/refund",
		"POST "+base+"/remove_pos_orders/:pos_id",
	)
	assert.ElementsMatch(t, expected, registeredRoutes(t, orderRoutes(&v1.OrderRest{})))
}

func TestInvoiceIssueKeepsItsLegacyPath(t *testing.T) {
	expected := append(legacyBuiltinRoutes(models.InvoiceSchemaName),
		"POST "+routeV1Prefix+"/"+models.InvoiceSchemaName+"/:id/issue")
	assert.ElementsMatch(t, expected, registeredRoutes(t, invoiceRoutes(&v1.InvoiceRest{})))
}
