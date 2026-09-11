package composable

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"

	ds "github.com/sky-as-code/nikki-erp/common/datastructure"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
)

// ownerContext is an authenticated owner, who passes every permission check. The principal is
// not decoration: AssertPermission fails closed without one.
func ownerContext() corectx.Context {
	ctx := corectx.NewRequestContext(context.Background())
	ctx.SetPermissions(corectx.ContextPermissions{
		IsOwner:   true,
		Principal: corectx.Principal{Kind: corectx.PrincipalKindUser, Id: model.Id("owner_test")},
	})
	return ctx
}

// deniedContext fails every permission check.
func deniedContext() corectx.Context {
	ctx := corectx.NewRequestContext(context.Background())
	ctx.SetPermissions(corectx.ContextPermissions{})
	return ctx
}

// memberContext is a caller who belongs to org_mine and to nothing else.
func memberContext() corectx.Context {
	ctx := corectx.NewRequestContext(context.Background())
	ctx.SetPermissions(corectx.ContextPermissions{
		IsOwner:    true,
		Principal:  corectx.Principal{Kind: corectx.PrincipalKindUser, Id: model.Id("member_test")},
		UserOrgIds: ds.NewSetFrom(model.Id("org_mine")),
	})
	return ctx
}

// serviceContext is a service principal minted for one org. It holds no org membership, which
// is the whole reason org scope cannot be resolved from UserOrgIds alone.
func serviceContext(orgId *model.Id) corectx.Context {
	ctx := corectx.NewRequestContext(context.Background())
	ctx.SetPermissions(corectx.ContextPermissions{
		IsOwner: true,
		Principal: corectx.Principal{
			Kind:  corectx.PrincipalKindService,
			Id:    model.Id("svc_test"),
			OrgId: orgId,
		},
	})
	return ctx
}

// newPlainSchema declares no org column, so the org-scoping machinery leaves it alone.
func newPlainSchema() *dmodel.ModelSchema {
	return dmodel.DefineModel("cmp_plain_resource").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeString(0, 50))).
		Field(dmodel.DefineField().Name("name").DataType(dmodel.FieldDataTypeString(0, 200))).
		Build()
}

// newOrgSchema declares an org column, which is what switches org scoping on. The keys are
// plain strings so a test can use readable ids without tripping ULID validation.
func newOrgSchema() *dmodel.ModelSchema {
	return dmodel.DefineModel("cmp_org_resource").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeString(0, 50))).
		Field(dmodel.DefineField().Name(basemodel.FieldOrgId).DataType(dmodel.FieldDataTypeString(0, 50))).
		Field(dmodel.DefineField().Name("name").DataType(dmodel.FieldDataTypeString(0, 200))).
		Field(dmodel.DefineField().Name(basemodel.FieldIsArchived).DataType(dmodel.FieldDataTypeBoolean())).
		Build()
}

// newOrgRequiredSchema declares org_id as required-for-create, which is how the real org-scoped
// resources declare it: a create that omits it must be reported by schema validation.
func newOrgRequiredSchema() *dmodel.ModelSchema {
	return dmodel.DefineModel("cmp_org_required_resource").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeString(0, 50))).
		Field(dmodel.DefineField().Name(basemodel.FieldOrgId).
			DataType(dmodel.FieldDataTypeString(0, 50)).IsRequiredForCreate(true)).
		Field(dmodel.DefineField().Name("name").DataType(dmodel.FieldDataTypeString(0, 200))).
		Build()
}

// stubRepository records the keys it was asked to fetch and returns a canned record. For the
// bulk paths it also opens a fake transaction and answers Search from searchItems.
type stubRepository struct {
	CrudRepository
	schema      *dmodel.ModelSchema
	fetchedKeys dmodel.DynamicFields
	record      dmodel.DynamicFields
	found       bool

	searchItems []dmodel.DynamicFields
	searchGraph *dmodel.SearchGraph
	tranx       *stubTransaction
}

type stubTransaction struct {
	committed  bool
	rolledBack bool
}

func (this *stubTransaction) Commit() error   { this.committed = true; return nil }
func (this *stubTransaction) Rollback() error { this.rolledBack = true; return nil }

func (this *stubRepository) Schema() *dmodel.ModelSchema { return this.schema }

func (this *stubRepository) BeginTransaction(_ corectx.Context) (database.DbTransaction, error) {
	this.tranx = &stubTransaction{}
	return this.tranx, nil
}

func (this *stubRepository) Search(_ corectx.Context, param dyn.RepoSearchParam) (*SearchResult, error) {
	this.searchGraph = param.Graph
	return &SearchResult{
		Data:    dyn.PagedResultData[dmodel.DynamicFields]{Items: this.searchItems, Total: len(this.searchItems)},
		HasData: len(this.searchItems) > 0,
	}, nil
}

func (this *stubRepository) FindByKeys(
	_ corectx.Context, keys dmodel.DynamicFields,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	this.fetchedKeys = keys
	return &dyn.OpResult[dmodel.DynamicFields]{Data: this.record, HasData: this.found}, nil
}

// fakeDomainService records what reaches it and answers success, so the application layer can
// be exercised without a database.
type fakeDomainService struct {
	CrudDomainService
	schema  *dmodel.ModelSchema
	repo    CrudRepository
	seen    dmodel.DynamicFields
	history []dmodel.DynamicFields
	calls   int

	// rejectName makes Create/Update answer a client error for a record with that name, so the
	// bulk paths can be shown skipping a bad row without a database.
	rejectName string

	// idPrefix, when set, makes Create answer a generated id ("{prefix}{n}") like a real create.
	idPrefix string

	createOpts CreateOptions
	updateOpts UpdateOptions
	deleteOpts DeleteOptions
}

func newFakeDomainService(schema *dmodel.ModelSchema) *fakeDomainService {
	return &fakeDomainService{schema: schema, repo: &stubRepository{schema: schema}}
}

func (this *fakeDomainService) Schema() *dmodel.ModelSchema { return this.schema }
func (this *fakeDomainService) Repository() CrudRepository  { return this.repo }

func (this *fakeDomainService) record(params dmodel.DynamicFields) {
	this.calls++
	this.seen = params
	this.history = append(this.history, params)
}

func (this *fakeDomainService) rejects(cmd dmodel.DynamicFields) ft.ClientErrors {
	if this.rejectName == "" || readString(cmd, "name") != this.rejectName {
		return nil
	}
	return ft.ClientErrors{*ft.NewValidationError("name", "err_test_rejected", "rejected by the fake")}
}

func (this *fakeDomainService) Create(_ corectx.Context, cmd CreateCommand, options ...CreateOptions) (*CreateResult, error) {
	this.record(cmd)
	this.createOpts = firstOrZero(options)
	if cErrs := this.rejects(cmd); cErrs != nil {
		return &CreateResult{ClientErrors: cErrs}, nil
	}
	data := cmd
	if this.idPrefix != "" {
		data = copyFields(cmd)
		data[basemodel.FieldId] = this.idPrefix + strconv.Itoa(this.calls)
	}
	return &CreateResult{Data: data, HasData: true}, nil
}

func (this *fakeDomainService) Update(_ corectx.Context, cmd UpdateCommand, options ...UpdateOptions) (*MutateResult, error) {
	this.record(cmd)
	this.updateOpts = firstOrZero(options)
	if cErrs := this.rejects(cmd); cErrs != nil {
		return &MutateResult{ClientErrors: cErrs}, nil
	}
	return &MutateResult{HasData: true}, nil
}

func (this *fakeDomainService) Delete(_ corectx.Context, cmd DeleteCommand, options ...DeleteOptions) (*MutateResult, error) {
	this.record(cmd)
	this.deleteOpts = firstOrZero(options)
	return &MutateResult{HasData: true}, nil
}

func (this *fakeDomainService) SetArchived(_ corectx.Context, cmd SetArchivedCommand) (*MutateResult, error) {
	this.record(cmd)
	return &MutateResult{HasData: true}, nil
}

func (this *fakeDomainService) GetById(_ corectx.Context, query GetByIdQuery) (*GetOneResult, error) {
	this.record(query)
	return &GetOneResult{Data: dyn.SingleResultData[dmodel.DynamicFields]{Item: dmodel.DynamicFields{"id": query["id"]}}, HasData: true}, nil
}

func (this *fakeDomainService) GetOne(_ corectx.Context, query GetOneQuery) (*GetOneResult, error) {
	this.record(query)
	return &GetOneResult{Data: dyn.SingleResultData[dmodel.DynamicFields]{Item: query}, HasData: true}, nil
}

func (this *fakeDomainService) Search(_ corectx.Context, query SearchQuery, _ ...corecrud.ServiceSearchOptions) (*SearchResult, error) {
	this.record(query)
	return &SearchResult{Data: dyn.PagedResultData[dmodel.DynamicFields]{Items: []dmodel.DynamicFields{}}, HasData: true}, nil
}

func (this *fakeDomainService) Exists(_ corectx.Context, query ExistsQuery) (*ExistsResult, error) {
	this.record(query)
	return &ExistsResult{HasData: true}, nil
}

func (this *fakeDomainService) ComputeField(_ corectx.Context, cmd ComputeFieldCommand) (*ComputeFieldResult, error) {
	this.record(cmd)
	return &ComputeFieldResult{Data: ComputeFieldResultData{Value: "computed"}, HasData: true}, nil
}

// newAppService wires the default application service over a fake domain service.
func newAppService(schema *dmodel.ModelSchema, param ...NewAppServiceParam) (*fakeDomainService, CrudApplicationService) {
	domSvc := newFakeDomainService(schema)
	p := firstOrZero(param)
	p.DomainService = domSvc
	return domSvc, NewDefaultApplicationService(p)
}

// stubHandlers is a CrudRestHandlers whose every handler answers its own name, so a route test
// can tell which one echo dispatched to.
type stubHandlers struct {
	appSvc CrudApplicationService
}

func (this *stubHandlers) answer(name string) echo.HandlerFunc {
	return func(echoCtx *echo.Context) error { return echoCtx.String(http.StatusOK, name) }
}

func (this *stubHandlers) Create(echoCtx *echo.Context) error { return this.answer("create")(echoCtx) }
func (this *stubHandlers) Update(echoCtx *echo.Context) error { return this.answer("update")(echoCtx) }
func (this *stubHandlers) Delete(echoCtx *echo.Context) error { return this.answer("delete")(echoCtx) }
func (this *stubHandlers) SetArchived(echoCtx *echo.Context) error {
	return this.answer("set_archived")(echoCtx)
}
func (this *stubHandlers) GetById(echoCtx *echo.Context) error {
	return this.answer("get_by_id")(echoCtx)
}
func (this *stubHandlers) Search(echoCtx *echo.Context) error { return this.answer("search")(echoCtx) }
func (this *stubHandlers) Exists(echoCtx *echo.Context) error { return this.answer("exists")(echoCtx) }
func (this *stubHandlers) GetSchema(echoCtx *echo.Context) error {
	return this.answer("get_schema")(echoCtx)
}
func (this *stubHandlers) ComputeField(echoCtx *echo.Context) error {
	return this.answer("compute_field")(echoCtx)
}
func (this *stubHandlers) CreateBulk(echoCtx *echo.Context) error {
	return this.answer("bulk_create")(echoCtx)
}
func (this *stubHandlers) Import(echoCtx *echo.Context) error {
	return this.answer("import")(echoCtx)
}
func (this *stubHandlers) ApplicationService() CrudApplicationService {
	return this.appSvc
}

// registeredRoutes registers the engine onto a throwaway echo instance and returns the routes
// as "METHOD path", in registration order.
func registeredRoutes(t *testing.T, restEngine RestEngine) []string {
	t.Helper()

	echoApp := echo.New()
	if err := restEngine.RegisterRoutes(echoApp.Group(""), passThrough); err != nil {
		t.Fatalf("RegisterRoutes: %v", err)
	}
	routes := echoApp.Router().Routes()
	result := make([]string, 0, len(routes))
	for _, route := range routes {
		result = append(result, route.Method+" "+route.Path)
	}
	return result
}

func passThrough(next echo.HandlerFunc) echo.HandlerFunc {
	return next
}

// installContext is the group middleware that does what the request-context middleware does
// in production: every handler and every route middleware finds a corectx on the request.
func installContext(ctx corectx.Context) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(echoCtx *echo.Context) error {
			echoCtx.SetRequest(echoCtx.Request().WithContext(ctx))
			return next(echoCtx)
		}
	}
}

// serveWithContext registers the engine behind a request context and serves one request.
func serveWithContext(
	t *testing.T, restEngine RestEngine, ctx corectx.Context, request *http.Request,
) *httptest.ResponseRecorder {
	t.Helper()

	echoApp := echo.New()
	if err := restEngine.RegisterRoutes(echoApp.Group("", installContext(ctx)), passThrough); err != nil {
		t.Fatalf("RegisterRoutes: %v", err)
	}
	recorder := httptest.NewRecorder()
	echoApp.ServeHTTP(recorder, request)
	return recorder
}

// jsonRequest is a request carrying a JSON body, which is what a write action binds from.
func jsonRequest(method string, target string, body string) *http.Request {
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return request
}

func indexOf(values []string, target string) int {
	for index, value := range values {
		if value == target {
			return index
		}
	}
	return -1
}
