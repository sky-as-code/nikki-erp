package engine

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	it "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// serveWithContext registers the engine's REST surface behind a middleware that installs a
// request context, which is what the real smoke-authorize middleware does in production.
func serveWithContext(t *testing.T, engine *DynamicResourceEngineImpl, request *http.Request) *httptest.ResponseRecorder {
	t.Helper()

	echoApp := echo.New()
	engine.RestApi().RegisterRoutes(echoApp.Group(""), func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(echoCtx *echo.Context) error {
			// AsRequestContext type-asserts the request's context to corectx.Context, so the
			// context itself is what has to be installed.
			echoCtx.SetRequest(echoCtx.Request().WithContext(memberContext()))
			return next(echoCtx)
		}
	})

	recorder := httptest.NewRecorder()
	echoApp.ServeHTTP(recorder, request)
	return recorder
}

// restEngine is an org-bearing engine exposing one GET action that reports what it received.
func restEngine(t *testing.T, seen *dmodel.DynamicFields, nested bool) *DynamicResourceEngineImpl {
	t.Helper()

	engine := newOrgScopedTestEngine()
	definition := it.DynamicActionDefinition{
		ActionName: "custom",
		ActionType: it.ActionTypeRead,
		RestPath:   "custom",
		MainProcess: func(_ corectx.Context, input it.ProcessInput) (*it.ActionResult, error) {
			*seen = input.Params
			return &it.ActionResult{HasData: true, Data: "ok"}, nil
		},
	}
	if nested {
		definition.PrimarySchema = util.ToPtr("vdmc_kiosks")
		definition.PrimaryRestIdParam = util.ToPtr("kiosk_id")
	}
	require.NoError(t, engine.DefineAction(definition))
	return engine
}

func TestRestRejectsAMissingOrgIdWith400(t *testing.T) {
	var seen dmodel.DynamicFields
	engine := restEngine(t, &seen, false)

	recorder := serveWithContext(t, engine,
		httptest.NewRequest(http.MethodGet, "/test_org_resource/custom", nil))

	// A caller who forgot the org gets a 400 naming the field, not a 500 and not a silent
	// unscoped read.
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), basemodel.FieldOrgId)
	assert.Nil(t, seen)
}

func TestRestAcceptsSnakeCaseOrgIdQueryParam(t *testing.T) {
	var seen dmodel.DynamicFields
	engine := restEngine(t, &seen, false)

	recorder := serveWithContext(t, engine,
		httptest.NewRequest(http.MethodGet, "/test_org_resource/custom?org_id=org_mine", nil))

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "org_mine", seen[basemodel.FieldOrgId])
}

func TestRestNoLongerAcceptsTheOldCamelCaseSpelling(t *testing.T) {
	var seen dmodel.DynamicFields
	engine := restEngine(t, &seen, false)

	// "orgId" was the old spelling. It is now just an unrecognised query param, so the
	// request is missing its org and is refused.
	recorder := serveWithContext(t, engine,
		httptest.NewRequest(http.MethodGet, "/test_org_resource/custom?orgId=org_mine", nil))

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestRestRefusesAnOrgTheCallerDoesNotBelongTo(t *testing.T) {
	var seen dmodel.DynamicFields
	engine := restEngine(t, &seen, false)

	recorder := serveWithContext(t, engine,
		httptest.NewRequest(http.MethodGet, "/test_org_resource/custom?org_id=org_theirs", nil))

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Nil(t, seen)
}

func TestNestedRouteResolvesAndBindsTheParentId(t *testing.T) {
	var seen dmodel.DynamicFields
	engine := restEngine(t, &seen, true)

	recorder := serveWithContext(t, engine, httptest.NewRequest(
		http.MethodGet,
		"/vdmc_kiosks/kiosk_7/test_org_resource/custom?org_id=org_mine",
		nil,
	))

	require.Equal(t, http.StatusOK, recorder.Code)
	// The parent id reaches the action under the name the definition declared, so a module
	// can filter on it without the engine having to know what it means.
	assert.Equal(t, "kiosk_7", seen["kiosk_id"])
}

func TestFlatRouteIsGoneOnceAnActionNests(t *testing.T) {
	var seen dmodel.DynamicFields
	engine := restEngine(t, &seen, true)

	recorder := serveWithContext(t, engine,
		httptest.NewRequest(http.MethodGet, "/test_org_resource/custom?org_id=org_mine", nil))

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

// writeRestEngine is an org-bearing engine exposing one write-typed action that reports what
// it received, so a test can assert where the org was read from.
func writeRestEngine(
	t *testing.T, seen *dmodel.DynamicFields, actionType it.ActionType,
) *DynamicResourceEngineImpl {
	t.Helper()

	engine := newOrgScopedTestEngine()
	require.NoError(t, engine.DefineAction(it.DynamicActionDefinition{
		ActionName: "custom",
		ActionType: actionType,
		RestPath:   "custom",
		MainProcess: func(_ corectx.Context, input it.ProcessInput) (*it.ActionResult, error) {
			*seen = input.Params
			return &it.ActionResult{HasData: true, Data: "ok"}, nil
		},
	}))
	return engine
}

// jsonRequest is a request carrying a JSON body, which is what a write action binds from.
func jsonRequest(method string, target string, body string) *http.Request {
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return request
}

func TestRestReadsOrgIdFromTheBodyOnWriteActions(t *testing.T) {
	// The four body-carrying types all take their org from the record they carry, so none of
	// these needs the org repeated in the URL.
	for _, actionType := range []it.ActionType{
		it.ActionTypeCreate,
		it.ActionTypeUpdatePatch,
		it.ActionTypeUpdateReplace,
		it.ActionTypeGeneric,
	} {
		t.Run(string(actionType), func(t *testing.T) {
			var seen dmodel.DynamicFields
			engine := writeRestEngine(t, &seen, actionType)

			recorder := serveWithContext(t, engine, jsonRequest(
				actionType.HttpMethod(),
				"/test_org_resource/custom",
				`{"org_id":"org_mine"}`,
			))

			require.Equal(t, http.StatusOK, recorder.Code)
			assert.Equal(t, "org_mine", seen[basemodel.FieldOrgId])
		})
	}
}

func TestRestPrefersTheBodyOrgOverTheQueryParamOnWrites(t *testing.T) {
	var seen dmodel.DynamicFields
	engine := writeRestEngine(t, &seen, it.ActionTypeCreate)

	// The record decides which org it belongs to. A stale or wrong org in the URL must not
	// silently redirect the write somewhere the payload never named.
	recorder := serveWithContext(t, engine, jsonRequest(
		http.MethodPost,
		"/test_org_resource/custom?org_id=org_theirs",
		`{"org_id":"org_mine"}`,
	))

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "org_mine", seen[basemodel.FieldOrgId])
}

func TestRestFallsBackToTheQueryParamForABodylessWrite(t *testing.T) {
	var seen dmodel.DynamicFields
	engine := writeRestEngine(t, &seen, it.ActionTypeGeneric)

	// A generic action may be a bodyless POST such as ":id/confirm", which has no body to read
	// an org from. Dropping the query fallback would make those unreachable.
	recorder := serveWithContext(t, engine,
		httptest.NewRequest(http.MethodPost, "/test_org_resource/custom?org_id=org_mine", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "org_mine", seen[basemodel.FieldOrgId])
}

func TestRestStillRejectsAWriteThatNamesNoOrgAtAll(t *testing.T) {
	var seen dmodel.DynamicFields
	engine := writeRestEngine(t, &seen, it.ActionTypeCreate)

	recorder := serveWithContext(t, engine,
		jsonRequest(http.MethodPost, "/test_org_resource/custom", `{}`))

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), basemodel.FieldOrgId)
	assert.Nil(t, seen)
}

func TestRestRefusesABodyOrgTheCallerDoesNotBelongTo(t *testing.T) {
	var seen dmodel.DynamicFields
	engine := writeRestEngine(t, &seen, it.ActionTypeCreate)

	// Membership is checked wherever the org came from: moving the value into the body must
	// not become a way around the check.
	recorder := serveWithContext(t, engine, jsonRequest(
		http.MethodPost, "/test_org_resource/custom", `{"org_id":"org_theirs"}`))

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Nil(t, seen)
}

func TestRestStillRequiresTheQueryParamOnReadActions(t *testing.T) {
	var seen dmodel.DynamicFields
	engine := restEngine(t, &seen, false)

	// A read is unchanged by the body-org rule: mergeOrgId does not treat it as a body action,
	// so with no "?org_id=" there is nothing to bind and the request is refused.
	recorder := serveWithContext(t, engine,
		httptest.NewRequest(http.MethodGet, "/test_org_resource/custom", nil))

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Nil(t, seen)
}

func TestIsBodyOrgActionClassifiesByActionType(t *testing.T) {
	var seen dmodel.DynamicFields

	for _, actionType := range []it.ActionType{
		it.ActionTypeCreate,
		it.ActionTypeUpdatePatch,
		it.ActionTypeUpdateReplace,
		it.ActionTypeGeneric,
	} {
		restApi := writeRestEngine(t, &seen, actionType).RestApi().(*DynamicRestApiImpl)
		assert.True(t, restApi.isBodyOrgAction("custom"), string(actionType))
	}

	for _, actionType := range []it.ActionType{it.ActionTypeRead, it.ActionTypeDelete} {
		restApi := writeRestEngine(t, &seen, actionType).RestApi().(*DynamicRestApiImpl)
		assert.False(t, restApi.isBodyOrgAction("custom"), string(actionType))
	}
}

func TestIsBodyOrgActionIsFalseForAnUnknownAction(t *testing.T) {
	var seen dmodel.DynamicFields
	restApi := writeRestEngine(t, &seen, it.ActionTypeCreate).RestApi().(*DynamicRestApiImpl)

	// Query-only is the safe direction: an action registered outside the normal path keeps
	// the behaviour it had before rather than erroring.
	assert.False(t, restApi.isBodyOrgAction("never_defined"))
}

func TestOrgLessResourceIgnoresAnOrgFromEitherSource(t *testing.T) {
	var seen dmodel.DynamicFields
	engine := newTestEngine()
	require.NoError(t, engine.DefineAction(it.DynamicActionDefinition{
		ActionName: "custom",
		ActionType: it.ActionTypeCreate,
		RestPath:   "custom",
		MainProcess: func(_ corectx.Context, input it.ProcessInput) (*it.ActionResult, error) {
			seen = input.Params
			return &it.ActionResult{HasData: true, Data: "ok"}, nil
		},
	}))

	recorder := serveWithContext(t, engine, jsonRequest(
		http.MethodPost,
		"/"+engine.RoutePath()+"/custom?org_id=org_mine",
		`{}`,
	))

	// The resource has no org column, so schemaHasOrgId gates mergeOrgId off: the query param
	// is never copied into params, and the request is not refused for lacking an org.
	require.Equal(t, http.StatusOK, recorder.Code)
	assert.NotContains(t, seen, basemodel.FieldOrgId)
}
