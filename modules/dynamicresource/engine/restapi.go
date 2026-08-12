package engine

import (
	"sort"
	"strings"

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

// RegisterRoutes adds every REST-exposed action of the resource to the given group.
// An action opts in by declaring an ActionType, which decides the HTTP method, and a
// RestPath relative to the resource base.
//
// The order matters: echo matches in registration order, so the literal "meta/schema" and
// "exists" paths must be registered before the ":id" patterns that would swallow them.
// routableActions sorts by specificity so that holds without hand-written ordering.
func (this *DynamicRestApiImpl) RegisterRoutes(route *echo.Group, middlewares ...echo.MiddlewareFunc) {
	base := "/" + this.engine.RoutePath()

	for _, definition := range this.routableActions() {
		handler := definition.RestHandler
		if handler == nil {
			handler = this.actionHandler(definition.ActionName)
		}
		route.Add(
			definition.ActionType.HttpMethod(),
			joinRestPath(base, definition.RestPath),
			handler,
			middlewares...,
		)
	}
}

// routableActions returns the actions that declare a REST surface, ordered by decreasing
// specificity: paths with fewer path params first, then longer paths, then alphabetically.
// That places "meta/schema" and "exists" ahead of ":id", and ":id/archived" ahead of ":id".
func (this *DynamicRestApiImpl) routableActions() []it.DynamicActionDefinition {
	definitions := make([]it.DynamicActionDefinition, 0, len(this.engine.ActionNames()))
	for _, name := range this.engine.ActionNames() {
		definition, exists := this.engine.Action(name)
		if exists && definition.ActionType != "" {
			definitions = append(definitions, definition)
		}
	}

	sort.SliceStable(definitions, func(i, j int) bool {
		left, right := definitions[i].RestPath, definitions[j].RestPath
		if leftParams, rightParams := countPathParams(left), countPathParams(right); leftParams != rightParams {
			return leftParams < rightParams
		}
		if leftDepth, rightDepth := pathDepth(left), pathDepth(right); leftDepth != rightDepth {
			return leftDepth > rightDepth
		}
		if left != right {
			return left < right
		}
		// Two actions on the same path differ only by HTTP method and cannot shadow each
		// other; order them by name so registration stays deterministic.
		return definitions[i].ActionName < definitions[j].ActionName
	})
	return definitions
}

// actionHandler builds the generic handler for an action, resolving the request and response
// shaping through the built-in binding table.
func (this *DynamicRestApiImpl) actionHandler(actionName string) echo.HandlerFunc {
	binding := this.bindingFor(actionName)
	return func(echoCtx *echo.Context) error {
		return this.serveAction(
			echoCtx, actionName, binding.buildParams, binding.buildResponse, binding.jsonSuccess,
		)
	}
}

// restBinding is how one action turns an HTTP request into action params and an action
// result back into an HTTP payload.
type restBinding struct {
	buildParams   buildParamsFn
	buildResponse buildResponseFn
	jsonSuccess   func(*echo.Context, any) error
}

// bindingFor returns the built-in binding of an action, or the generic one for any action a
// feature module defined. The generic binding assembles params the way echo.Bind does and
// passes the action result through untouched.
func (this *DynamicRestApiImpl) bindingFor(actionName string) restBinding {
	switch actionName {
	case it.ActionCreate:
		return restBinding{this.createBodyParams, createResponse, httpserver.JsonCreated}
	case it.ActionUpdate:
		return restBinding{this.updateParams, mutateResponse, httpserver.JsonOk}
	case it.ActionDelete:
		return restBinding{this.deleteParams, mutateResponse, httpserver.JsonOk}
	case it.ActionSetArchived:
		return restBinding{this.archivedParams, mutateResponse, httpserver.JsonOk}
	case it.ActionGetById:
		return restBinding{this.getByIdParams, getOneResponse, httpserver.JsonOk}
	case it.ActionSearch:
		return restBinding{this.searchParams, searchResponse, httpserver.JsonOk}
	case it.ActionExists:
		return restBinding{rawBodyParams, identityResponse, httpserver.JsonOk}
	case it.ActionGetSchema:
		return restBinding{noParams, identityResponse, httpserver.JsonOk}
	}
	return restBinding{echoBindParams, identityResponse, httpserver.JsonOk}
}

// joinRestPath places a RestPath under the resource base. An empty RestPath is the base itself.
func joinRestPath(base string, restPath string) string {
	if restPath == "" {
		return base
	}
	return base + "/" + restPath
}

func countPathParams(restPath string) int {
	return strings.Count(restPath, ":")
}

func pathDepth(restPath string) int {
	if restPath == "" {
		return 0
	}
	return strings.Count(restPath, "/") + 1
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
		// A well-formed body naming fields the schema does not declare is the caller's
		// mistake, not a server fault.
		if unknownFields, isUnknown := err.(*unknownFieldsError); isUnknown {
			return httpserver.JsonBadRequest(echoCtx, unknownFields.errors)
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
	// Search reports HasData=false for an empty page, which is a successful result rather
	// than a missing record: a filter matching nothing, or a page past the end, still
	// answers 200 with an empty item list.
	if !result.HasData && actionName != it.ActionSearch {
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
