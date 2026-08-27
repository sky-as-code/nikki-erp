package engine

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"

	it "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// registeredRoutes registers the engine's REST surface onto a throwaway echo instance and
// returns the routes as "METHOD path", in registration order.
func registeredRoutes(t *testing.T, engine *DynamicResourceEngineImpl) []string {
	t.Helper()

	echoApp := echo.New()
	engine.RestApi().RegisterRoutes(echoApp.Group(""))

	routes := echoApp.Router().Routes()
	result := make([]string, 0, len(routes))
	for _, route := range routes {
		result = append(result, route.Method+" "+route.Path)
	}
	return result
}

// The built-in routes must come out byte-identical to the hand-written ones this replaced,
// and in an order where literal segments precede the ":id" patterns that would swallow them.
func TestRegisterRoutesCoversBuiltins(t *testing.T) {
	engine := newTestEngine()
	assert.NoError(t, DefineBuiltinActions(engine))

	// Within one path, order is by action name ("create" before "search"); different HTTP
	// methods on the same path cannot shadow each other, so that tie is cosmetic.
	assert.Equal(t, []string{
		"GET /test_resource/meta/schema",
		"POST /test_resource/exists",
		"POST /test_resource",
		"GET /test_resource",
		// A literal "meta/compute" segment must sort ahead of ":id", or the id pattern would
		// swallow it.
		"POST /test_resource/meta/compute/:field",
		"POST /test_resource/:id/archived",
		"DELETE /test_resource/:id",
		"GET /test_resource/:id",
		"PATCH /test_resource/:id",
	}, registeredRoutes(t, engine))
}

// get_by_unique declares no ActionType, so it must not reach the REST surface.
func TestRegisterRoutesSkipsUnexposedActions(t *testing.T) {
	engine := newTestEngine()
	assert.NoError(t, DefineBuiltinActions(engine))

	for _, route := range registeredRoutes(t, engine) {
		assert.NotContains(t, route, "get_by_unique")
	}
	assert.Len(t, registeredRoutes(t, engine), 9, "10 built-ins, 1 unexposed")
}

// A module-defined action gets a route from its RestPath, which is the whole point of the
// uplift: before it, only the built-ins were reachable over HTTP.
func TestRegisterRoutesIncludesCustomAction(t *testing.T) {
	engine := newTestEngine()
	assert.NoError(t, DefineBuiltinActions(engine))
	assert.NoError(t, engine.DefineAction(it.DynamicActionDefinition{
		ActionName:  "send_invitation",
		ActionType:  it.ActionTypeGeneric,
		RestPath:    ":id/send_invitation",
		MainProcess: noopProcess,
	}))

	routes := registeredRoutes(t, engine)
	assert.Contains(t, routes, "POST /test_resource/:id/send_invitation")

	// More specific than the bare ":id" patterns, so it must be registered before them.
	assert.Less(t,
		indexOf(routes, "POST /test_resource/:id/send_invitation"),
		indexOf(routes, "PATCH /test_resource/:id"),
	)
}

// A RestHandler replaces the generic handler entirely, owning request and response shaping.
func TestRegisterRoutesUsesCustomRestHandler(t *testing.T) {
	engine := newTestEngine()
	assert.NoError(t, engine.DefineAction(it.DynamicActionDefinition{
		ActionName: "custom",
		ActionType: it.ActionTypeRead,
		RestPath:   "custom",
		RestHandler: func(echoCtx *echo.Context) error {
			return echoCtx.String(http.StatusTeapot, "brewed")
		},
		MainProcess: noopProcess,
	}))

	echoApp := echo.New()
	engine.RestApi().RegisterRoutes(echoApp.Group(""))

	recorder := httptest.NewRecorder()
	echoApp.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/test_resource/custom", nil))

	assert.Equal(t, http.StatusTeapot, recorder.Code, "the custom handler ran, not the pipeline")
	assert.Equal(t, "brewed", recorder.Body.String())
}

func TestJoinRestPath(t *testing.T) {
	assert.Equal(t, "/thing", joinRestPath("/thing", ""))
	assert.Equal(t, "/thing/:id", joinRestPath("/thing", ":id"))
	assert.Equal(t, "/thing/meta/schema", joinRestPath("/thing", "meta/schema"))
}

// echoBindParams binds path params, then query params for GET/DELETE only, then the body,
// each step overwriting the previous — the precedence echo.Bind documents.
func TestEchoBindParamsPrecedence(t *testing.T) {
	echoApp := echo.New()
	echoApp.GET("/thing/:id", func(echoCtx *echo.Context) error {
		params, err := echoBindParams(echoCtx)
		assert.NoError(t, err)
		assert.Equal(t, "from_path", params["id"])
		assert.Equal(t, "from_query", params["note"])
		return echoCtx.NoContent(http.StatusOK)
	})

	recorder := httptest.NewRecorder()
	echoApp.ServeHTTP(recorder, httptest.NewRequest(
		http.MethodGet, "/thing/from_path?note=from_query", nil,
	))
	assert.Equal(t, http.StatusOK, recorder.Code)
}

// On a POST the query string is not bound, matching echo.Bind, and the body wins over the path.
func TestEchoBindParamsBodyOverridesPath(t *testing.T) {
	echoApp := echo.New()
	echoApp.POST("/thing/:id", func(echoCtx *echo.Context) error {
		params, err := echoBindParams(echoCtx)
		assert.NoError(t, err)
		assert.Equal(t, "from_body", params["id"])
		assert.NotContains(t, params, "note", "query params are not bound on POST")
		return echoCtx.NoContent(http.StatusOK)
	})

	request := httptest.NewRequest(
		http.MethodPost, "/thing/from_path?note=ignored", strings.NewReader(`{"id":"from_body"}`),
	)
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	recorder := httptest.NewRecorder()
	echoApp.ServeHTTP(recorder, request)
	assert.Equal(t, http.StatusOK, recorder.Code)
}

// A bodyless POST must not fail on an EOF from the JSON decoder.
func TestEchoBindParamsToleratesEmptyBody(t *testing.T) {
	echoApp := echo.New()
	echoApp.POST("/thing/:id", func(echoCtx *echo.Context) error {
		params, err := echoBindParams(echoCtx)
		assert.NoError(t, err)
		assert.Equal(t, "from_path", params["id"])
		return echoCtx.NoContent(http.StatusOK)
	})

	recorder := httptest.NewRecorder()
	echoApp.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/thing/from_path", nil))
	assert.Equal(t, http.StatusOK, recorder.Code)
}

func indexOf(values []string, target string) int {
	for index, value := range values {
		if value == target {
			return index
		}
	}
	return -1
}
