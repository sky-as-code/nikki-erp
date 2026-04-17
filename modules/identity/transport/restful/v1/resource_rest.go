package v1

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/resource"
)

type resourceRestParams struct {
	dig.In

	ResourceSvc it.ResourceService
}

func NewResourceRest(params resourceRestParams) *ResourceRest {
	return &ResourceRest{ResourceSvc: params.ResourceSvc}
}

type ResourceRest struct {
	httpserver.RestBase
	ResourceSvc it.ResourceService
}

func (this ResourceRest) CreateResource(echoCtx *echo.Context) (err error) {
	return httpserver.ServeCreate(
		"create resource",
		echoCtx,
		&it.CreateResourceCommand{},
		this.ResourceSvc.CreateResource,
	)
}

func (this ResourceRest) DeleteResource(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"delete resource",
		echoCtx,
		this.ResourceSvc.DeleteResource,
	)
}

func (this ResourceRest) GetResource(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne(
		"get resource",
		echoCtx,
		this.ResourceSvc.GetResource,
	)
}

func (this ResourceRest) ResourceExists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeExists(
		"resource exists",
		echoCtx,
		this.ResourceSvc.ResourceExists,
	)
}

func (this ResourceRest) SearchResources(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch(
		"search resources",
		echoCtx,
		this.ResourceSvc.SearchResources,
	)
}

func (this ResourceRest) UpdateResource(echoCtx *echo.Context) (err error) {
	return httpserver.ServeUpdate(
		"update resource",
		echoCtx,
		&it.UpdateResourceCommand{},
		this.ResourceSvc.UpdateResource,
	)
}

/*
 * Non-CRUD APIs
 */

func (this ResourceRest) GetModelSchema(echoCtx *echo.Context) (err error) {
	schema := dmodel.MustGetSchema(domain.ResourceSchemaName)
	echoCtx.JSON(http.StatusOK, schema.ToSimplized())
	return nil
}
