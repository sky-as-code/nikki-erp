package engine

import (
	"github.com/labstack/echo/v5"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	it "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

func NewDynamicRestApi(engine it.DynamicResourceEngine) it.DynamicRestApi {
	return &DynamicRestApiImpl{engine: engine}
}

// DynamicRestApiImpl exposes one resource over HTTP. Every endpoint goes through
// Engine.ExecuteAction rather than calling the service directly, so that a permission
// check always runs and a ModifyAction override takes effect on the REST surface too.
type DynamicRestApiImpl struct {
	engine it.DynamicResourceEngine
}

// RegisterRoutes adds every endpoint of the resource to the given group.
// The order matters: echo matches in registration order, so the literal "meta/schema"
// and "exists" paths must be registered before the ":id" patterns that would swallow them.
func (this *DynamicRestApiImpl) RegisterRoutes(route *echo.Group, middlewares ...echo.MiddlewareFunc) {
	base := "/" + this.engine.RoutePath()

	route.DELETE(base+"/:id", this.deleteResource, middlewares...)
	route.GET(base+"/meta/schema", this.getModelSchema, middlewares...)
	route.GET(base+"/:id", this.getResourceById, middlewares...)
	route.GET(base, this.searchResources, middlewares...)
	route.POST(base, this.createResource, middlewares...)
	route.POST(base+"/exists", this.resourceExists, middlewares...)
	route.POST(base+"/:id/archived", this.setResourceArchived, middlewares...)
	route.POST(base+"/:id", this.updateResource, middlewares...)
}

func (this *DynamicRestApiImpl) deleteResource(echoCtx *echo.Context) error {
	return this.serveAction(echoCtx, it.ActionDelete, this.deleteParams, mutateResponse, httpserver.JsonOk)
}

func (this *DynamicRestApiImpl) getModelSchema(echoCtx *echo.Context) error {
	return this.serveAction(echoCtx, it.ActionGetSchema, noParams, identityResponse, httpserver.JsonOk)
}

func (this *DynamicRestApiImpl) getResourceById(echoCtx *echo.Context) error {
	return this.serveAction(echoCtx, it.ActionGetById, this.getByIdParams, getOneResponse, httpserver.JsonOk)
}

func (this *DynamicRestApiImpl) searchResources(echoCtx *echo.Context) error {
	return this.serveAction(echoCtx, it.ActionSearch, this.searchParams, searchResponse, httpserver.JsonOk)
}

func (this *DynamicRestApiImpl) createResource(echoCtx *echo.Context) error {
	return this.serveAction(echoCtx, it.ActionCreate, this.bodyParams, createResponse, httpserver.JsonCreated)
}

func (this *DynamicRestApiImpl) resourceExists(echoCtx *echo.Context) error {
	return this.serveAction(echoCtx, it.ActionExists, rawBodyParams, identityResponse, httpserver.JsonOk)
}

func (this *DynamicRestApiImpl) setResourceArchived(echoCtx *echo.Context) error {
	return this.serveAction(echoCtx, it.ActionSetArchived, this.archivedParams, mutateResponse, httpserver.JsonOk)
}

func (this *DynamicRestApiImpl) updateResource(echoCtx *echo.Context) error {
	return this.serveAction(echoCtx, it.ActionUpdate, this.updateParams, mutateResponse, httpserver.JsonOk)
}

// buildParamsFn assembles the action params out of the HTTP request.
type buildParamsFn func(echoCtx *echo.Context) (dmodel.DynamicFields, error)

// buildResponseFn reshapes the action result data into the HTTP payload.
type buildResponseFn func(data any) any

// serveAction is the single path every endpoint takes. It mirrors the error handling of
// httpserver.ServeRequestDynamic: a malformed body or a client error answers 400, a
// missing record answers 400 with a not-found payload, and a Go error bubbles up as 500.
func (this *DynamicRestApiImpl) serveAction(
	echoCtx *echo.Context,
	actionName string,
	buildParams buildParamsFn,
	buildResponse buildResponseFn,
	jsonSuccessFn func(*echo.Context, any) error,
) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST "+actionName); e != nil {
			err = e
		}
	}()

	reqCtx, err := corectx.AsRequestContext(echoCtx)
	if err != nil {
		return err
	}

	params, err := buildParams(echoCtx)
	if err != nil {
		if _, isHttpErr := err.(*echo.HTTPError); isHttpErr {
			return httpserver.JsonBadRequest(echoCtx, []any{
				ft.NewAnonymousValidationError(ft.ErrorKey("err_malformed_request"), "malformed request"),
			})
		}
		return err
	}

	result, err := this.engine.ExecuteAction(reqCtx, actionName, params)
	if err != nil {
		return err
	}

	if result.ClientErrors != nil && result.ClientErrors.Count() > 0 {
		return httpserver.JsonBadRequest(echoCtx, result.ClientErrors)
	}
	if !result.HasData {
		return httpserver.JsonBadRequest(echoCtx, ft.ClientErrors{*ft.NewAnonymousNotFoundError()})
	}

	return jsonSuccessFn(echoCtx, buildResponse(result.Data))
}

// The response builders below type-assert the action result to the shape its built-in
// action documents. A custom action that returns something else should install its own
// REST surface rather than reusing these endpoints.

func identityResponse(data any) any {
	return data
}

func createResponse(data any) any {
	fields, ok := data.(dmodel.DynamicFields)
	if !ok {
		return data
	}
	return httpserver.NewRestCreateResponseDyn(fields)
}

func mutateResponse(data any) any {
	mutation, ok := data.(dyn.MutateResultData)
	if !ok {
		return data
	}
	return httpserver.NewRestMutateResponse(mutation)
}

func getOneResponse(data any) any {
	single, ok := data.(dyn.SingleResultData[dmodel.DynamicFields])
	if !ok {
		return data
	}
	return httpserver.RestGetOneResponse[dmodel.DynamicFields]{
		Item: single.Item,
		Meta: single.Meta,
	}
}

func searchResponse(data any) any {
	paged, ok := data.(dyn.PagedResultData[dmodel.DynamicFields])
	if !ok {
		return data
	}
	return httpserver.RestSearchResponse[dmodel.DynamicFields]{
		Items:         paged.Items,
		Total:         paged.Total,
		Page:          paged.Page,
		Size:          paged.Size,
		DesiredFields: paged.DesiredFields,
		MaskedFields:  paged.MaskedFields,
		SchemaEtag:    paged.SchemaEtag,
	}
}
