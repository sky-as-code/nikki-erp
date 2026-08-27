package engine

import (
	"net/http"
	"net/http/httptest"
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
