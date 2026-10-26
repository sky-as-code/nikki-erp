package restful

import (
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/settings/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/settings/dynamicengines"
	v1 "github.com/sky-as-code/nikki-erp/modules/settings/transport/restful/v1"
)

func TestEngineNamesMatchTheRegisteredOnions(t *testing.T) {
	assert.Equal(t, composable.EngineDependencyName(models.SettingsSchemaSchemaName), v1.SettingsSchemaEngineName)
	assert.Equal(t, composable.EngineDependencyName(models.SettingsRecordSchemaName), v1.SettingsRecordEngineName)
	assert.Equal(t, composable.EngineDependencyName(models.SettingsQuotaSchemaName), v1.SettingsQuotaEngineName)
}

func passThrough(next echo.HandlerFunc) echo.HandlerFunc {
	return next
}

// Settings serves reads only: no create, update, delete, archive, bulk or import route may exist
// on either resource, because a module writes its own settings through PATCH
// /v1/{module_key}/settings/:level.
func TestSettingsResourcesServeReadsOnly(t *testing.T) {
	for schemaName, handlers := range map[string]composable.CrudRestHandlers{
		models.SettingsSchemaSchemaName: &v1.SettingsSchemaRest{},
		models.SettingsRecordSchemaName: &v1.SettingsRecordRest{},
		models.SettingsQuotaSchemaName:  &v1.SettingsQuotaRest{},
	} {
		t.Run(schemaName, func(t *testing.T) {
			echoApp := echo.New()
			require.NoError(t, composable.NewRestEngine(schemaName, handlers).
				AddCrudRoutes(dynamicengines.ReadOnlyCrudActions()...).
				RegisterRoutes(echoApp.Group("/v1/settings"), passThrough))

			got := []string{}
			for _, route := range echoApp.Router().Routes() {
				got = append(got, route.Method+" "+route.Path)
			}
			base := "/v1/settings/" + schemaName
			assert.Equal(t, []string{
				"GET " + base + "/meta/schema",
				"POST " + base + "/exists",
				"GET " + base,
				"POST " + base + "/meta/compute/:field",
				"GET " + base + "/:id",
			}, got)
		})
	}
}
