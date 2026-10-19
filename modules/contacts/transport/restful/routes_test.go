package restful

import (
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/modules/contacts/domain/models"
	v1 "github.com/sky-as-code/nikki-erp/modules/contacts/transport/restful/v1"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The container names in the REST structs' tags are literals; they must be what the onions are
// registered under, or the handlers resolve nothing at boot.
func TestEngineNamesMatchTheRegisteredOnions(t *testing.T) {
	assert.Equal(t, composable.EngineDependencyName(models.PartySchemaName), v1.PartyEngineName)
	assert.Equal(t, composable.EngineDependencyName(models.CommChannelSchemaName), v1.CommChannelEngineName)
	assert.Equal(t, composable.EngineDependencyName(models.RelationshipSchemaName), v1.RelationshipEngineName)
	assert.Equal(t, composable.EngineDependencyName(models.VendorProfileSchemaName), v1.VendorProfileEngineName)
}

func passThrough(next echo.HandlerFunc) echo.HandlerFunc {
	return next
}

// The legacy engine served these paths; the frontend and the Robot suites address them as-is.
func TestBuiltinRoutesOfEveryContactsResource(t *testing.T) {
	for schemaName, handlers := range map[string]composable.CrudRestHandlers{
		models.PartySchemaName:         &v1.PartyRest{},
		models.CommChannelSchemaName:   &v1.CommChannelRest{},
		models.RelationshipSchemaName:  &v1.RelationshipRest{},
		models.VendorProfileSchemaName: &v1.VendorProfileRest{},
	} {
		t.Run(schemaName, func(t *testing.T) {
			echoApp := echo.New()
			group := echoApp.Group("/v1/contacts")
			require.NoError(t, composable.NewRestEngine(schemaName, handlers).AddCrudRoutes().
				RegisterRoutes(group, passThrough))

			got := []string{}
			for _, route := range echoApp.Router().Routes() {
				got = append(got, route.Method+" "+route.Path)
			}
			base := "/v1/contacts/" + schemaName
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
			}, got)
		})
	}
}
