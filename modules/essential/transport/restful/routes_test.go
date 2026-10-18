package restful

import (
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/essential/dynamicengines"
	v1 "github.com/sky-as-code/nikki-erp/modules/essential/transport/restful/v1"
)

// The container names in the REST structs' tags are literals; they must be what the onions are
// registered under, or the handlers resolve nothing at boot.
func TestEngineNamesMatchTheRegisteredOnions(t *testing.T) {
	assert.Equal(t, composable.EngineDependencyName(models.CurrencySchemaName), v1.CurrencyEngineName)
	assert.Equal(t, composable.EngineDependencyName(models.UomSchemaName), v1.UomEngineName)
	assert.Equal(t, composable.EngineDependencyName(models.UomCatSchemaName), v1.UomCatEngineName)
	assert.Equal(t, composable.EngineDependencyName(models.LanguageSchemaName), v1.LanguageEngineName)
	assert.Equal(t, composable.EngineDependencyName(models.EnumSchemaName), v1.EnumEngineName)
	assert.Equal(t, composable.EngineDependencyName(models.TagSchemaName), v1.TagEngineName)
	assert.Equal(t, composable.EngineDependencyName(models.FieldMetadataSchemaName), v1.FieldMetadataEngineName)
	assert.Equal(t, composable.EngineDependencyName(models.ModelMetadataSchemaName), v1.ModelMetadataEngineName)
	assert.Equal(t, composable.EngineDependencyName(models.ModuleMetadataSchemaName), v1.ModuleMetadataEngineName)
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
		models.CurrencySchemaName:      &v1.CurrencyRest{},
		models.UomSchemaName:           &v1.UomRest{},
		models.UomCatSchemaName:        &v1.UomCatRest{},
		models.EnumSchemaName:          &v1.EnumRest{},
		models.TagSchemaName:           &v1.TagRest{},
		models.FieldMetadataSchemaName: &v1.FieldMetadataRest{},
		models.ModelMetadataSchemaName: &v1.ModelMetadataRest{},
	} {
		t.Run(schemaName, func(t *testing.T) {
			base := "/v1/essential/" + schemaName
			assert.Equal(t, []string{
				"GET " + base + "/meta/schema",
				"POST " + base + "/bulk",
				"POST " + base + "/exists",
				"POST " + base + "/import",
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

// The two public language reads are collection actions; their literal segments must register
// ahead of ":id", and they must not carry the authorizing middleware (the sign-in page reads them).
func TestLanguageCustomRoutes(t *testing.T) {
	echoApp := echo.New()
	rest := &v1.LanguageRest{}
	noop := func(*echo.Context, map[string]any) error { return nil }
	restEngine := composable.NewRestEngine(models.LanguageSchemaName, rest).
		AddCrudRoutes().
		AddRoute(composable.RouteDefinition{Path: "json", ActionType: composable.ActionTypeRead, HandlerFn: noop}).
		AddRoute(composable.RouteDefinition{Path: "enabled", ActionType: composable.ActionTypeRead, HandlerFn: noop})
	require.NoError(t, restEngine.RegisterRoutes(echoApp.Group("/v1/essential"), passThrough))

	got := map[string]int{}
	for i, route := range echoApp.Router().Routes() {
		got[route.Method+" "+route.Path] = i
	}
	base := "/v1/essential/" + models.LanguageSchemaName
	for _, literal := range []string{base + "/json", base + "/enabled"} {
		position, ok := got["GET "+literal]
		require.True(t, ok, literal)
		assert.Less(t, position, got["GET "+base+"/:id"], literal+" must precede :id")
	}
}

// The module catalogue is read-only over REST: its rows belong to the boot-time sync.
func TestModuleCatalogueServesReadsOnly(t *testing.T) {
	echoApp := echo.New()
	require.NoError(t, composable.NewRestEngine(models.ModuleMetadataSchemaName, &v1.ModuleRest{}).
		AddCrudRoutes(dynamicengines.ModuleReadActions()...).
		RegisterRoutes(echoApp.Group("/v1/essential"), passThrough))

	got := []string{}
	for _, route := range echoApp.Router().Routes() {
		got = append(got, route.Method+" "+route.Path)
	}
	base := "/v1/essential/" + models.ModuleMetadataSchemaName
	assert.Equal(t, []string{
		"GET " + base + "/meta/schema",
		"POST " + base + "/exists",
		"GET " + base,
		"GET " + base + "/:id",
	}, got)
}
