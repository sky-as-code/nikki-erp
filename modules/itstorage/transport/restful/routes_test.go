package restful

import (
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/domain/models"
	v1 "github.com/sky-as-code/nikki-erp/modules/itstorage/transport/restful/v1"
)

func TestEngineNamesMatchTheRegisteredOnions(t *testing.T) {
	assert.Equal(t, composable.EngineDependencyName(models.SpaceSchemaName), v1.SpaceEngineName)
	assert.Equal(t, composable.EngineDependencyName(models.FileSchemaName), v1.FileEngineName)
}

// The BFF codes against these three paths; they must not drift under the engine's resource paths.
func TestFileRoutesMatchTheContract(t *testing.T) {
	echoApp := echo.New()
	registerFileRoutes(echoApp.Group("/v1/itstorage"), &v1.FileRest{})

	got := []string{}
	for _, route := range echoApp.Router().Routes() {
		got = append(got, route.Method+" "+route.Path)
	}
	assert.ElementsMatch(t, []string{
		"POST /v1/itstorage/files/upload",
		"GET /v1/itstorage/files/meta",
		"GET /v1/itstorage/files/download",
	}, got)
}
