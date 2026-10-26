package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	m "github.com/sky-as-code/nikki-erp/modules/core/httpserver/middlewares"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	c "github.com/sky-as-code/nikki-erp/modules/itstorage/constants"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/dynamicengines"
	v1 "github.com/sky-as-code/nikki-erp/modules/itstorage/transport/restful/v1"
)

func InitRestfulHandlers() error {
	return stdErr.Join(
		deps.Register(v1.NewSpaceRest, v1.NewFileRest, v1.NewConsumptionRest),
		initItStorageV1(),
	)
}

func initItStorageV1() error {
	return deps.Invoke(func(
		route *echo.Group, spaceRest *v1.SpaceRest, fileRest *v1.FileRest, consumptionRest *v1.ConsumptionRest,
	) error {
		routeV1 := route.Group("/v1/itstorage")
		registerFileRoutes(routeV1, fileRest)
		return stdErr.Join(
			composable.NewRestEngine(models.SpaceSchemaName, spaceRest).
				AddCrudRoutes(dynamicengines.SpaceRestActions()...).
				RegisterRoutes(routeV1),
			composable.NewRestEngine(models.FileSchemaName, fileRest).
				AddCrudRoutes(dynamicengines.FileRestActions()...).
				RegisterRoutes(routeV1),
			composable.NewRestEngine(models.ConsumptionSchemaName, consumptionRest).
				AddCrudRoutes(dynamicengines.ConsumptionRestActions()...).
				RegisterRoutes(routeV1),
		)
	})
}

// registerFileRoutes mounts the contract the Web BFF codes against (constants.Route*). They are
// hand-written rather than engine routes: the engine would bind and buffer the multipart body,
// and it answers JSON where these stream.
func registerFileRoutes(routeV1 *echo.Group, fileRest *v1.FileRest) {
	routeV1.POST(c.RouteFileUpload, fileRest.Upload, m.UserSmokeAuthz())
	routeV1.GET(c.RouteFileMeta, fileRest.GetMeta, m.UserSmokeAuthz())
	routeV1.GET(c.RouteFileDownload, fileRest.Download, m.UserSmokeAuthz())
	routeV1.GET(c.RouteFileDownloadById, fileRest.DownloadById, m.UserSmokeAuthz())
	routeV1.GET(c.RouteFileMetaById, fileRest.GetMetaById, m.UserSmokeAuthz())
}
