package composable

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// newCrudRest wires the real CrudRestBase over the default application service and a fake
// domain service, so the binders are exercised end to end without a database.
func newCrudRest(schema *dmodel.ModelSchema) (*fakeDomainService, RestEngine) {
	domSvc, appSvc := newAppService(schema)
	rest := &CrudRestBase{}
	rest.SetApplicationService(appSvc)
	return domSvc, NewRestEngine(schema.Name(), rest).AddCrudRoutes()
}

func TestCreateRejectsUnknownBodyFieldsWith400(t *testing.T) {
	domSvc, restEngine := newCrudRest(newPlainSchema())

	recorder := serveWithContext(t, restEngine, ownerContext(),
		jsonRequest(http.MethodPost, "/cmp_plain_resource", `{"name":"x","bogus":1}`))

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "bogus")
	assert.Zero(t, domSvc.calls, "a request naming undeclared fields never reaches the service")
}

func TestCreateAnswers201WithDeclaredFields(t *testing.T) {
	domSvc, restEngine := newCrudRest(newPlainSchema())

	recorder := serveWithContext(t, restEngine, ownerContext(),
		jsonRequest(http.MethodPost, "/cmp_plain_resource", `{"name":"x"}`))

	assert.Equal(t, http.StatusCreated, recorder.Code)
	assert.Equal(t, "x", domSvc.seen["name"])
}

// Update keeps the permissive binding: clients round-tripping a fetched record back may carry
// keys they never intended to write.
func TestUpdateDropsUnknownFieldsAndTakesTheIdFromThePath(t *testing.T) {
	domSvc, restEngine := newCrudRest(newPlainSchema())

	recorder := serveWithContext(t, restEngine, ownerContext(),
		jsonRequest(http.MethodPatch, "/cmp_plain_resource/rec_1", `{"id":"other","name":"y","bogus":1}`))

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "rec_1", domSvc.seen["id"], "the route decides which record is updated")
	assert.Equal(t, "y", domSvc.seen["name"])
	assert.NotContains(t, domSvc.seen, "bogus")
}

func TestSearchParsesPagingFieldsGraphAndArchivedFlag(t *testing.T) {
	domSvc, restEngine := newCrudRest(newPlainSchema())

	recorder := serveWithContext(t, restEngine, ownerContext(), httptest.NewRequest(http.MethodGet,
		"/cmp_plain_resource?page=2&size=10&fields=id,name&include_archived=true&graph=%7B%22and%22%3A%5B%5D%7D", nil))

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, 2, domSvc.seen["page"])
	assert.Equal(t, 10, domSvc.seen["size"])
	assert.Equal(t, []string{"id", "name"}, domSvc.seen["fields"])
	assert.Equal(t, true, domSvc.seen["include_archived"])
	assert.NotNil(t, domSvc.seen["graph"])
}

func TestSearchRejectsAMalformedGraphWith400(t *testing.T) {
	domSvc, restEngine := newCrudRest(newPlainSchema())

	recorder := serveWithContext(t, restEngine, ownerContext(),
		httptest.NewRequest(http.MethodGet, "/cmp_plain_resource?graph=not-json", nil))

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Zero(t, domSvc.calls)
}

func TestReadsTakeTheOrgFromTheQueryString(t *testing.T) {
	domSvc, restEngine := newCrudRest(newOrgSchema())
	domSvc.repo.(*stubRepository).found = true

	for name, request := range map[string]*http.Request{
		"search":    httptest.NewRequest(http.MethodGet, "/cmp_org_resource?org_id=org_mine", nil),
		"get_by_id": httptest.NewRequest(http.MethodGet, "/cmp_org_resource/rec_1?org_id=org_mine", nil),
		"delete":    httptest.NewRequest(http.MethodDelete, "/cmp_org_resource/rec_1?org_id=org_mine", nil),
	} {
		t.Run(name, func(t *testing.T) {
			recorder := serveWithContext(t, restEngine, memberContext(), request)
			assert.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
			assert.Equal(t, "org_mine", domSvc.seen[basemodel.FieldOrgId])
		})
	}
}

func TestReadRejectsAMissingOrgWith400(t *testing.T) {
	domSvc, restEngine := newCrudRest(newOrgSchema())

	recorder := serveWithContext(t, restEngine, memberContext(),
		httptest.NewRequest(http.MethodGet, "/cmp_org_resource", nil))

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), basemodel.FieldOrgId)
	assert.Zero(t, domSvc.calls)
}

// A write action carries the record, and the record carries its own org.
func TestCreateTakesTheOrgFromTheBodyAndBodyWinsOverQuery(t *testing.T) {
	domSvc, restEngine := newCrudRest(newOrgSchema())

	recorder := serveWithContext(t, restEngine, memberContext(),
		jsonRequest(http.MethodPost, "/cmp_org_resource?org_id=org_someone_else", `{"org_id":"org_mine","name":"x"}`))

	assert.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())
	assert.Equal(t, "org_mine", domSvc.seen[basemodel.FieldOrgId])
}

func TestCreateFallsBackToTheQueryOrgWhenTheBodyNamesNone(t *testing.T) {
	domSvc, restEngine := newCrudRest(newOrgSchema())

	recorder := serveWithContext(t, restEngine, memberContext(),
		jsonRequest(http.MethodPost, "/cmp_org_resource?org_id=org_mine", `{"name":"x"}`))

	assert.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())
	assert.Equal(t, "org_mine", domSvc.seen[basemodel.FieldOrgId])
}

// An org-less resource never sees an invented org_id key, or create's unknown-field check would
// refuse a perfectly valid body.
func TestOrgLessResourceIgnoresAnOrgQueryParam(t *testing.T) {
	domSvc, restEngine := newCrudRest(newPlainSchema())

	recorder := serveWithContext(t, restEngine, memberContext(),
		jsonRequest(http.MethodPost, "/cmp_plain_resource?org_id=org_mine", `{"name":"x"}`))

	assert.Equal(t, http.StatusCreated, recorder.Code)
	assert.NotContains(t, domSvc.seen, basemodel.FieldOrgId)
}

func TestSchemaEndpointNeedsNoOrg(t *testing.T) {
	_, restEngine := newCrudRest(newOrgSchema())

	recorder := serveWithContext(t, restEngine, memberContext(),
		httptest.NewRequest(http.MethodGet, "/cmp_org_resource/meta/schema", nil))

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
}

func TestSearchAnswers200ForAnEmptyPage(t *testing.T) {
	_, restEngine := newCrudRest(newPlainSchema())

	recorder := serveWithContext(t, restEngine, ownerContext(),
		httptest.NewRequest(http.MethodGet, "/cmp_plain_resource", nil))

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"items":[]`)
}

func TestSetArchivedReadsTheFlagFromTheBody(t *testing.T) {
	domSvc, restEngine := newCrudRest(newPlainSchema())

	recorder := serveWithContext(t, restEngine, ownerContext(),
		jsonRequest(http.MethodPost, "/cmp_plain_resource/rec_1/archived", `{"is_archived":true,"etag":"e1"}`))

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "rec_1", domSvc.seen["id"])
	assert.Equal(t, true, domSvc.seen[basemodel.FieldIsArchived])
}
