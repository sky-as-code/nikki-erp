package composable

import (
	"github.com/labstack/echo/v5"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
)

// CrudRestHandlers is what RestEngine.AddCrudRoutes binds. CrudRestBase implements it; a module's
// {Resource}Rest embeds CrudRestBase and adds its own handlers for AddRoute.
type CrudRestHandlers interface {
	Create(echoCtx *echo.Context) error
	Update(echoCtx *echo.Context) error
	Delete(echoCtx *echo.Context) error
	SetArchived(echoCtx *echo.Context) error
	GetById(echoCtx *echo.Context) error
	Search(echoCtx *echo.Context) error
	Exists(echoCtx *echo.Context) error
	GetSchema(echoCtx *echo.Context) error
	ComputeField(echoCtx *echo.Context) error
	CreateBulk(echoCtx *echo.Context) error
	Import(echoCtx *echo.Context) error

	ApplicationService() CrudApplicationService
}

// CrudRestBase serves the built-in actions of one resource over HTTP. Every handler binds the
// request, calls the application service (which authorizes) and shapes the response the way
// the hand-written REST handlers in modules/core/httpserver do.
//
//	type UomRest struct{ composable.CrudRestBase }
//
//	func NewUomRest(param struct {
//		dig.In
//		Engine composable.DynamicResourceEngineOnion `name:"dynengine_essential_uom"`
//	}) *UomRest {
//		rest := &UomRest{}
//		rest.SetApplicationService(param.Engine.ApplicationService())
//		return rest
//	}
type CrudRestBase struct {
	httpserver.RestBase
	appSvc CrudApplicationService
}

func (this *CrudRestBase) SetApplicationService(svc CrudApplicationService) {
	this.appSvc = svc
}

func (this *CrudRestBase) ApplicationService() CrudApplicationService {
	return this.appSvc
}

func (this *CrudRestBase) schema() *dmodel.ModelSchema {
	if this.appSvc == nil {
		return nil
	}
	return this.appSvc.Schema()
}

func (this *CrudRestBase) Create(echoCtx *echo.Context) error {
	return serve(echoCtx, "create",
		func(echoCtx *echo.Context) (dmodel.DynamicFields, error) {
			return createBodyParams(echoCtx, this.schema())
		},
		this.appSvc.Create, createResponse, httpserver.JsonCreated, false)
}

func (this *CrudRestBase) Update(echoCtx *echo.Context) error {
	return serve(echoCtx, "update",
		func(echoCtx *echo.Context) (dmodel.DynamicFields, error) { return updateParams(echoCtx, this.schema()) },
		this.appSvc.Update, mutateResponse, httpserver.JsonOk, false)
}

func (this *CrudRestBase) Delete(echoCtx *echo.Context) error {
	return serve(echoCtx, "delete",
		func(echoCtx *echo.Context) (dmodel.DynamicFields, error) { return deleteParams(echoCtx, this.schema()) },
		this.appSvc.Delete, mutateResponse, httpserver.JsonOk, false)
}

func (this *CrudRestBase) SetArchived(echoCtx *echo.Context) error {
	return serve(echoCtx, "set archived",
		func(echoCtx *echo.Context) (dmodel.DynamicFields, error) {
			return archivedParams(echoCtx, this.schema())
		},
		this.appSvc.SetArchived, mutateResponse, httpserver.JsonOk, false)
}

func (this *CrudRestBase) GetById(echoCtx *echo.Context) error {
	return serve(echoCtx, "get by id",
		func(echoCtx *echo.Context) (dmodel.DynamicFields, error) {
			return getByIdParams(echoCtx, this.schema())
		},
		this.appSvc.GetById, getOneResponse, httpserver.JsonOk, false)
}

// Search answers 200 with an empty item list for a filter matching nothing, or a page past the
// end: HasData=false is a successful result there, not a missing record.
func (this *CrudRestBase) Search(echoCtx *echo.Context) error {
	return serve(echoCtx, "search",
		func(echoCtx *echo.Context) (dmodel.DynamicFields, error) { return searchParams(echoCtx, this.schema()) },
		this.appSvc.Search, searchResponse, httpserver.JsonOk, true)
}

func (this *CrudRestBase) Exists(echoCtx *echo.Context) error {
	return serve(echoCtx, "exists",
		func(echoCtx *echo.Context) (dmodel.DynamicFields, error) {
			return bindRawBodyParams(echoCtx, this.schema(), ActionTypeGeneric)
		},
		this.appSvc.Exists, identity[dyn.ExistsResultData], httpserver.JsonOk, false)
}

func (this *CrudRestBase) GetSchema(echoCtx *echo.Context) error {
	return serve(echoCtx, "get schema", noParams, this.appSvc.GetSchema, identity[any], httpserver.JsonOk, false)
}

// ComputeField serves POST {resource}/meta/compute/:field with a body of {model, args}.
func (this *CrudRestBase) ComputeField(echoCtx *echo.Context) error {
	return serve(echoCtx, "compute field",
		func(echoCtx *echo.Context) (dmodel.DynamicFields, error) {
			return bindPayload(echoCtx, nil, ActionTypeGeneric)
		},
		this.appSvc.ComputeField, identity[ComputeFieldResultData], httpserver.JsonOk, false)
}

// CreateBulk serves POST {resource}/bulk with a body of {org_id?, items: [...]}. The body is
// bound raw: "items" is not a schema field, and each item is filtered by the create path.
func (this *CrudRestBase) CreateBulk(echoCtx *echo.Context) error {
	return serve(echoCtx, "create bulk",
		func(echoCtx *echo.Context) (dmodel.DynamicFields, error) {
			return bindRawBodyParams(echoCtx, this.schema(), ActionTypeCreate)
		},
		this.appSvc.CreateBulk, identity[BulkCreateResultData], httpserver.JsonOk, false)
}

// serve is the single path every built-in endpoint takes. It mirrors the error handling of
// httpserver.ServeRequestDynamic: a malformed body or a client error answers 400, a missing
// record answers 400 with a not-found payload, and a Go error bubbles up as 500.
func serve[TData any](
	echoCtx *echo.Context,
	action string,
	bind bindFn,
	call func(ctx corectx.Context, params dmodel.DynamicFields) (*dyn.OpResult[TData], error),
	shape func(data TData) any,
	jsonSuccess func(*echo.Context, any) error,
	emptyIsSuccess bool,
) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST "+action); e != nil {
			err = e
		}
	}()

	reqCtx, err := corectx.AsRequestContext(echoCtx)
	if err != nil {
		return err
	}

	params, err := bind(echoCtx)
	if err != nil {
		return bindFailure(echoCtx, err)
	}

	result, err := call(reqCtx, params)
	if err != nil {
		return err
	}
	if result.ClientErrors != nil && result.ClientErrors.Count() > 0 {
		return httpserver.JsonBadRequest(echoCtx, result.ClientErrors)
	}
	if !result.HasData && !emptyIsSuccess {
		return httpserver.JsonBadRequest(echoCtx, ft.ClientErrors{*ft.NewAnonymousNotFoundError()})
	}
	return jsonSuccess(echoCtx, shape(result.Data))
}

// bindFailure maps a binding error to the response the caller should see. A malformed body and
// a body naming undeclared fields are both the caller's mistake, not a server fault.
func bindFailure(echoCtx *echo.Context, err error) error {
	if _, isHttpErr := err.(*echo.HTTPError); isHttpErr {
		return httpserver.JsonBadRequest(echoCtx, []any{
			ft.NewAnonymousValidationError(ft.ErrorKey("err_malformed_request"), "malformed request"),
		})
	}
	if unknownFields, isUnknown := err.(*unknownFieldsError); isUnknown {
		return httpserver.JsonBadRequest(echoCtx, unknownFields.errors)
	}
	return err
}

func identity[TData any](data TData) any {
	return data
}

func createResponse(fields dmodel.DynamicFields) any {
	return httpserver.NewRestCreateResponseDyn(fields)
}

func mutateResponse(mutation dyn.MutateResultData) any {
	return httpserver.NewRestMutateResponse(mutation)
}

func getOneResponse(single dyn.SingleResultData[dmodel.DynamicFields]) any {
	single.Meta.DesiredFields = emptyIfNil(single.Meta.DesiredFields)
	single.Meta.MaskedFields = emptyIfNil(single.Meta.MaskedFields)
	return httpserver.RestGetOneResponse[dmodel.DynamicFields]{
		Item: single.Item,
		Meta: single.Meta,
	}
}

func searchResponse(paged dyn.PagedResultData[dmodel.DynamicFields]) any {
	return httpserver.RestSearchResponse[dmodel.DynamicFields]{
		Items:         paged.Items,
		Total:         paged.Total,
		Page:          paged.Page,
		Size:          paged.Size,
		DesiredFields: emptyIfNil(paged.DesiredFields),
		MaskedFields:  emptyIfNil(paged.MaskedFields),
		SchemaEtag:    paged.SchemaEtag,
	}
}

// emptyIfNil keeps `desired_fields` / `masked_fields` JSON arrays rather than `null`: clients
// declare both as plain arrays and index into them without a nil guard.
func emptyIfNil(fields []string) []string {
	if fields == nil {
		return []string{}
	}
	return fields
}

// RecordIdParam is the path parameter every single-row built-in route carries.
const RecordIdParam = basemodel.FieldId
