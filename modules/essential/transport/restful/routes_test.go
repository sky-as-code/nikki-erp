package restful

import (
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	v1 "github.com/sky-as-code/nikki-erp/modules/essential/transport/restful/v1"
)

// The container names in the REST structs' tags are literals; they must be what the onions are
// registered under, or the handlers resolve nothing at boot.
func TestEngineNamesMatchTheRegisteredOnions(t *testing.T) {
	assert.Equal(t, composable.EngineDependencyName(models.CurrencySchemaName), v1.CurrencyEngineName)
	assert.Equal(t, composable.EngineDependencyName(models.UomSchemaName), v1.UomEngineName)
	assert.Equal(t, composable.EngineDependencyName(models.UomCatSchemaName), v1.UomCatEngineName)
}

// registeredRoutes registers one resource's REST surface on a throwaway echo instance and
// returns "METHOD path" in registration order.
func registeredRoutes(t *testing.T, schemaName string, handlers composable.CrudRestHandlers) []string {
	t.Helper()

	echoApp := echo.New()
	restEngine := composable.NewRestEngine(schemaName, handlers).AddCrudRoutes()
	require.NoError(t, restEngine.RegisterRoutes(echoApp.Group("/v1/essential"), passThrough))

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

// The frontend addresses these resources at /v1/essential/{schema_name} with the built-in route
// table (PATCH for update, POST :id/archived, POST exists, GET meta/schema). The literal
// segments must register ahead of the ":id" patterns that would swallow them.
func TestBuiltinRoutesOfEveryComposableResource(t *testing.T) {
	for schemaName, handlers := range map[string]composable.CrudRestHandlers{
		models.CurrencySchemaName: &v1.CurrencyRest{},
		models.UomSchemaName:      &v1.UomRest{},
		models.UomCatSchemaName:   &v1.UomCatRest{},
	} {
		t.Run(schemaName, func(t *testing.T) {
			base := "/v1/essential/" + schemaName
			assert.Equal(t, []string{
				"GET " + base + "/meta/schema",
				"POST " + base + "/exists",
				"POST " + base,
				"GET " + base,
				"POST " + base + "/meta/compute/:field",
				"POST " + base + "/:id/archived",
				"DELETE " + base + "/:id",
				"GET " + base + "/:id",
				"PATCH " + base + "/:id",
			}, registeredRoutes(t, schemaName, handlers))
		})
	}
}
