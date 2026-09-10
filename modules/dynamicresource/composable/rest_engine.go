package composable

import (
	stdErr "errors"
	"sort"
	"strings"

	"github.com/labstack/echo/v5"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	m "github.com/sky-as-code/nikki-erp/modules/core/httpserver/middlewares"
)

// HandlerFn is the shape of a custom route's handler. payload is the request bound the way
// echo.Bind does (path, query on GET/DELETE, JSON body) plus the resolved org_id; it is nil for
// ActionTypeUpload, whose multipart form the handler reads itself.
type HandlerFn func(echoCtx *echo.Context, payload map[string]any) error

// RouteDefinition declares one custom route of a resource.
type RouteDefinition struct {
	// Path is relative to the resource base and may carry Echo path params (":id"). An empty
	// string is the resource base itself. Segments are [a-zA-Z0-9_]; hyphens are rejected.
	Path string

	// ActionType decides the HTTP method and how the payload is bound.
	ActionType ActionType

	HandlerFn HandlerFn

	// IsAuthorized adds the SmokeAuthz middleware, which verifies the caller's token and then
	// asserts that some layer actually performed a permission check. Nil means true. A public
	// endpoint (false) gets PublicUnauthorized instead; it must never reach an AssertAction.
	IsAuthorized *bool

	// Middlewares are appended after the authorization middleware.
	Middlewares []echo.MiddlewareFunc
}

// RestEngine registers the routes of one resource onto an echo group, in an order that keeps a
// literal path such as "meta/schema" ahead of the ":id" pattern that would swallow it.
type RestEngine interface {
	// AddCrudRoutes adds the built-in route table, or only the named actions when given.
	// GetByUnique has no route: it is the service-only read by unique key.
	AddCrudRoutes(only ...CrudAction) RestEngine

	AddRoute(def RouteDefinition) RestEngine

	// RegisterRoutes adds every route under "/{schema}" of the group. The optional middlewares
	// replace SmokeAuthz on authorized routes, which is what a test passes to avoid the
	// container; a public route keeps PublicUnauthorized regardless.
	RegisterRoutes(group *echo.Group, authorizedMiddlewares ...echo.MiddlewareFunc) error
}

func NewRestEngine(schemaName string, handlers CrudRestHandlers) RestEngine {
	return &RestEngineImpl{
		schemaName: schemaName,
		handlers:   handlers,
	}
}

type RestEngineImpl struct {
	schemaName string
	handlers   CrudRestHandlers
	routes     []restRoute
	errs       []error
}

type restRoute struct {
	name       string
	method     string
	path       string
	handler    echo.HandlerFunc
	authorized bool
	extra      []echo.MiddlewareFunc
}

type crudRoute struct {
	action     CrudAction
	actionType ActionType
	path       string
	handler    func() echo.HandlerFunc
}

// builtinRoutes is the route table every resource gets. compute_field is always present so a
// form can evaluate a function-kind field; without one the endpoint answers the same client
// error the schema check produces.
func (this *RestEngineImpl) builtinRoutes() []crudRoute {
	h := this.handlers
	return []crudRoute{
		{CrudActionCreate, ActionTypeCreate, "", func() echo.HandlerFunc { return h.Create }},
		{CrudActionUpdate, ActionTypeUpdatePatch, ":id", func() echo.HandlerFunc { return h.Update }},
		{CrudActionDelete, ActionTypeDelete, ":id", func() echo.HandlerFunc { return h.Delete }},
		{CrudActionSetArchived, ActionTypeGeneric, ":id/archived", func() echo.HandlerFunc { return h.SetArchived }},
		{CrudActionGetById, ActionTypeRead, ":id", func() echo.HandlerFunc { return h.GetById }},
		{CrudActionSearch, ActionTypeRead, "", func() echo.HandlerFunc { return h.Search }},
		{CrudActionExists, ActionTypeGeneric, "exists", func() echo.HandlerFunc { return h.Exists }},
		{CrudActionGetSchema, ActionTypeRead, "meta/schema", func() echo.HandlerFunc { return h.GetSchema }},
		{CrudActionComputeField, ActionTypeGeneric, "meta/compute/:field", func() echo.HandlerFunc { return h.ComputeField }},
	}
}

func (this *RestEngineImpl) AddCrudRoutes(only ...CrudAction) RestEngine {
	if this.handlers == nil {
		this.errs = append(this.errs, errors.Errorf("resource '%s': AddCrudRoutes needs CRUD handlers", this.schemaName))
		return this
	}
	wanted := map[CrudAction]bool{}
	for _, action := range only {
		wanted[action] = true
	}
	for _, route := range this.builtinRoutes() {
		if len(wanted) > 0 && !wanted[route.action] {
			continue
		}
		this.add(restRoute{
			name:       string(route.action),
			method:     route.actionType.HttpMethod(),
			path:       route.path,
			handler:    route.handler(),
			authorized: true,
		})
	}
	return this
}

func (this *RestEngineImpl) AddRoute(def RouteDefinition) RestEngine {
	if !def.ActionType.IsValid() {
		this.errs = append(this.errs, errors.Errorf(
			"resource '%s': route '%s' has invalid action type '%s'", this.schemaName, def.Path, def.ActionType))
		return this
	}
	if def.HandlerFn == nil {
		this.errs = append(this.errs, errors.Errorf(
			"resource '%s': route '%s' has no handler", this.schemaName, def.Path))
		return this
	}
	this.add(restRoute{
		name:       def.Path,
		method:     def.ActionType.HttpMethod(),
		path:       def.Path,
		handler:    this.wrapHandler(def),
		authorized: def.IsAuthorized == nil || *def.IsAuthorized,
		extra:      def.Middlewares,
	})
	return this
}

// wrapHandler binds the payload before handing the request over, so a custom handler never
// repeats the path/query/body/org assembly the built-in ones get for free.
func (this *RestEngineImpl) wrapHandler(def RouteDefinition) echo.HandlerFunc {
	return func(echoCtx *echo.Context) error {
		if def.ActionType == ActionTypeUpload {
			return def.HandlerFn(echoCtx, nil)
		}
		payload, err := bindPayload(echoCtx, this.schema(), def.ActionType)
		if err != nil {
			return bindFailure(echoCtx, err)
		}
		return def.HandlerFn(echoCtx, payload)
	}
}

func (this *RestEngineImpl) schema() *dmodel.ModelSchema {
	if this.handlers == nil || this.handlers.ApplicationService() == nil {
		return dmodel.GetSchema(this.schemaName)
	}
	return this.handlers.ApplicationService().Schema()
}

func (this *RestEngineImpl) add(route restRoute) {
	if route.path != "" && !RestPathRegex.MatchString(route.path) {
		this.errs = append(this.errs, errors.Errorf(
			"resource '%s': route path '%s' must be [a-zA-Z0-9_] segments, optionally ':' prefixed",
			this.schemaName, route.path))
		return
	}
	full := this.fullPath(route.path)
	for _, existing := range this.routes {
		if existing.method == route.method && this.fullPath(existing.path) == full {
			this.errs = append(this.errs, errors.Errorf(
				"resource '%s': route %s %s is already taken by '%s'",
				this.schemaName, route.method, full, existing.name))
			return
		}
	}
	this.routes = append(this.routes, route)
}

func (this *RestEngineImpl) fullPath(path string) string {
	base := "/" + this.schemaName
	if path == "" {
		return base
	}
	return base + "/" + path
}

// RegisterRoutes registers by decreasing specificity: paths with fewer path params first, then
// deeper paths, then alphabetically. Echo matches in registration order, so that places
// "meta/schema" and "exists" ahead of ":id", and ":id/archived" ahead of ":id".
func (this *RestEngineImpl) RegisterRoutes(group *echo.Group, authorizedMiddlewares ...echo.MiddlewareFunc) error {
	if err := stdErr.Join(this.errs...); err != nil {
		return err
	}
	if len(this.routes) == 0 {
		return errors.Errorf("resource '%s' declares no route", this.schemaName)
	}

	ordered := make([]restRoute, len(this.routes))
	copy(ordered, this.routes)
	sort.SliceStable(ordered, func(i, j int) bool {
		left, right := this.fullPath(ordered[i].path), this.fullPath(ordered[j].path)
		if leftParams, rightParams := countPathParams(left), countPathParams(right); leftParams != rightParams {
			return leftParams < rightParams
		}
		if leftDepth, rightDepth := pathDepth(left), pathDepth(right); leftDepth != rightDepth {
			return leftDepth > rightDepth
		}
		if left != right {
			return left < right
		}
		return ordered[i].name < ordered[j].name
	})

	authorized := authorizedMiddlewares
	if len(authorized) == 0 {
		authorized = []echo.MiddlewareFunc{m.SmokeAuthz()}
	}
	for _, route := range ordered {
		middlewares := []echo.MiddlewareFunc{m.PublicUnauthorized}
		if route.authorized {
			middlewares = authorized
		}
		group.Add(route.method, this.fullPath(route.path), route.handler, append(middlewares, route.extra...)...)
	}
	return nil
}

func countPathParams(path string) int {
	return strings.Count(path, ":")
}

func pathDepth(path string) int {
	if path == "" {
		return 0
	}
	return strings.Count(path, "/") + 1
}
