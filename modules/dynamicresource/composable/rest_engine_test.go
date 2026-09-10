package composable

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
)

func newRestEngine(schema *dmodel.ModelSchema) (RestEngine, *stubHandlers) {
	_, appSvc := newAppService(schema)
	handlers := &stubHandlers{appSvc: appSvc}
	return NewRestEngine(schema.Name(), handlers), handlers
}

// The built-in routes must come out byte-identical to the ones package engine registered, and
// in an order where literal segments precede the ":id" patterns that would swallow them.
func TestAddCrudRoutesCoversBuiltins(t *testing.T) {
	restEngine, _ := newRestEngine(newPlainSchema())
	restEngine.AddCrudRoutes()

	assert.Equal(t, []string{
		"GET /cmp_plain_resource/meta/schema",
		"POST /cmp_plain_resource/bulk",
		"POST /cmp_plain_resource/exists",
		"POST /cmp_plain_resource/import",
		"POST /cmp_plain_resource",
		"GET /cmp_plain_resource",
		"POST /cmp_plain_resource/meta/compute/:field",
		"POST /cmp_plain_resource/:id/archived",
		"DELETE /cmp_plain_resource/:id",
		"GET /cmp_plain_resource/:id",
		"PATCH /cmp_plain_resource/:id",
	}, registeredRoutes(t, restEngine))
}

func TestAddCrudRoutesHonoursTheOnlyFilter(t *testing.T) {
	restEngine, _ := newRestEngine(newPlainSchema())
	restEngine.AddCrudRoutes(CrudActionSearch, CrudActionGetById)

	assert.Equal(t, []string{
		"GET /cmp_plain_resource",
		"GET /cmp_plain_resource/:id",
	}, registeredRoutes(t, restEngine))
}

// A collection-level custom route is a literal segment, so it must register ahead of ":id".
func TestAddRouteSortsCustomRoutesBySpecificity(t *testing.T) {
	restEngine, _ := newRestEngine(newPlainSchema())
	restEngine.AddCrudRoutes().
		AddRoute(RouteDefinition{Path: "resolve_selection", ActionType: ActionTypeGeneric, HandlerFn: okHandler}).
		AddRoute(RouteDefinition{Path: ":id/suspend", ActionType: ActionTypeGeneric, HandlerFn: okHandler}).
		AddRoute(RouteDefinition{Path: ":id/effective", ActionType: ActionTypeRead, HandlerFn: okHandler})

	routes := registeredRoutes(t, restEngine)
	assert.Less(t, indexOf(routes, "POST /cmp_plain_resource/resolve_selection"), indexOf(routes, "GET /cmp_plain_resource/:id"))
	assert.Less(t, indexOf(routes, "POST /cmp_plain_resource/:id/suspend"), indexOf(routes, "PATCH /cmp_plain_resource/:id"))
	assert.Contains(t, routes, "GET /cmp_plain_resource/:id/effective")
}

func TestRegisterRoutesRefusesADuplicateRoute(t *testing.T) {
	restEngine, _ := newRestEngine(newPlainSchema())
	restEngine.AddCrudRoutes().
		AddRoute(RouteDefinition{Path: ":id", ActionType: ActionTypeRead, HandlerFn: okHandler})

	err := restEngine.RegisterRoutes(echo.New().Group(""), passThrough)

	require.Error(t, err, "two routes on one method and path would shadow each other silently")
	assert.Contains(t, err.Error(), "already taken")
}

func TestRegisterRoutesRefusesAHyphenatedPath(t *testing.T) {
	restEngine, _ := newRestEngine(newPlainSchema())
	restEngine.AddRoute(RouteDefinition{Path: "send-invitation", ActionType: ActionTypeGeneric, HandlerFn: okHandler})

	err := restEngine.RegisterRoutes(echo.New().Group(""), passThrough)

	require.Error(t, err)
}

func TestRegisterRoutesRefusesARouteWithoutHandler(t *testing.T) {
	restEngine, _ := newRestEngine(newPlainSchema())
	restEngine.AddRoute(RouteDefinition{Path: "custom", ActionType: ActionTypeGeneric})

	require.Error(t, restEngine.RegisterRoutes(echo.New().Group(""), passThrough))
}

// A custom handler receives the payload bound the way echo.Bind does, plus the org from the
// query string when the body named none.
func TestAddRouteBindsThePayloadForTheHandler(t *testing.T) {
	restEngine, _ := newRestEngine(newOrgSchema())
	var seen map[string]any
	restEngine.AddRoute(RouteDefinition{
		Path: ":id/suspend", ActionType: ActionTypeGeneric,
		HandlerFn: func(echoCtx *echo.Context, payload map[string]any) error {
			seen = payload
			return echoCtx.NoContent(http.StatusOK)
		},
	})

	recorder := serveWithContext(t, restEngine, memberContext(),
		jsonRequest(http.MethodPost, "/cmp_org_resource/rec_1/suspend?org_id=org_mine", `{"reason":"maintenance"}`))

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "rec_1", seen["id"])
	assert.Equal(t, "maintenance", seen["reason"])
	assert.Equal(t, "org_mine", seen["org_id"])
}

// An upload handler reads the multipart form itself, so the engine binds nothing.
func TestAddRouteSkipsBindingForUpload(t *testing.T) {
	restEngine, _ := newRestEngine(newOrgSchema())
	called := false
	restEngine.AddRoute(RouteDefinition{
		Path: ":id/upload", ActionType: ActionTypeUpload,
		HandlerFn: func(echoCtx *echo.Context, payload map[string]any) error {
			called = true
			assert.Nil(t, payload)
			return echoCtx.NoContent(http.StatusOK)
		},
	})

	recorder := serveWithContext(t, restEngine, memberContext(),
		httptest.NewRequest(http.MethodPost, "/cmp_org_resource/rec_1/upload", nil))

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.True(t, called)
}

// The authorized middleware guards every route except one declared public.
func TestRegisterRoutesAppliesAuthorizationPerRoute(t *testing.T) {
	restEngine, _ := newRestEngine(newPlainSchema())
	restEngine.
		AddRoute(RouteDefinition{Path: "private", ActionType: ActionTypeRead, HandlerFn: okHandler}).
		AddRoute(RouteDefinition{Path: "public", ActionType: ActionTypeRead, HandlerFn: okHandler, IsAuthorized: util.ToPtr(false)})

	guarded := []string{}
	recording := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(echoCtx *echo.Context) error {
			guarded = append(guarded, echoCtx.Request().URL.Path)
			return next(echoCtx)
		}
	}
	echoApp := echo.New()
	require.NoError(t, restEngine.RegisterRoutes(echoApp.Group("", installContext(memberContext())), recording))

	for _, path := range []string{"/cmp_plain_resource/private", "/cmp_plain_resource/public"} {
		recorder := httptest.NewRecorder()
		echoApp.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		assert.Equal(t, http.StatusOK, recorder.Code, path)
	}
	assert.Equal(t, []string{"/cmp_plain_resource/private"}, guarded,
		"only the authorized route passes through the authorization middleware")
}

func okHandler(echoCtx *echo.Context, _ map[string]any) error {
	return echoCtx.NoContent(http.StatusOK)
}
