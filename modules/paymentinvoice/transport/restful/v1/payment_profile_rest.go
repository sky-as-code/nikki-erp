package v1

import (
	"sort"

	"github.com/labstack/echo/v5"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
)

// The payment profile's own binding for create and update.
//
// The built-in binding filters a request body down to the schema fields, which is right for every
// other resource and wrong for this one: "config" is deliberately absent from the schema (that is
// what keeps the credentials out of a column), so the built-in binding would drop it on update and
// reject the request outright on create. Everything after binding is the application service's:
// the permission check, the org scope and the encrypting domain service run unchanged.

func (this *PaymentProfileRest) Create(echoCtx *echo.Context) error {
	return this.serveProfileWrite(echoCtx, true)
}

func (this *PaymentProfileRest) Update(echoCtx *echo.Context) error {
	return this.serveProfileWrite(echoCtx, false)
}

func (this *PaymentProfileRest) serveProfileWrite(echoCtx *echo.Context, isCreate bool) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST payment profile write"); e != nil {
			err = e
		}
	}()
	reqCtx, err := corectx.AsRequestContext(echoCtx)
	if err != nil {
		return err
	}
	params, cErrs := paymentProfileWriteParams(echoCtx, this.ApplicationService().Schema(), isCreate)
	if cErrs != nil {
		return httpserver.JsonBadRequest(echoCtx, *cErrs)
	}
	if isCreate {
		result, err := this.ApplicationService().Create(reqCtx, params)
		return answerProfileWrite(echoCtx, result, err, func(data dmodel.DynamicFields) error {
			return httpserver.JsonCreated(echoCtx, httpserver.NewRestCreateResponseDyn(data))
		})
	}
	result, err := this.ApplicationService().Update(reqCtx, params)
	return answerProfileWrite(echoCtx, result, err, func(data dyn.MutateResultData) error {
		return httpserver.JsonOk(echoCtx, httpserver.NewRestMutateResponse(data))
	})
}

// answerProfileWrite shapes the result exactly as the built-in handlers do, so a client cannot tell
// the two write endpoints apart from the rest of the module's CRUD.
func answerProfileWrite[TData any](
	echoCtx *echo.Context, result *dyn.OpResult[TData], err error, success func(TData) error,
) error {
	if err != nil {
		return err
	}
	if result.ClientErrors.Count() > 0 {
		return httpserver.JsonBadRequest(echoCtx, result.ClientErrors)
	}
	if !result.HasData {
		return httpserver.JsonBadRequest(echoCtx, ft.ClientErrors{*ft.NewAnonymousNotFoundError()})
	}
	return success(result.Data)
}

// paymentProfileWriteParams binds the body as the built-in binders would, then puts "config" back.
//
// Create rejects a body field that names nothing, matching the built-in create; update stays
// permissive, because clients round-trip a fetched record and carry read-only keys. The route, not
// the body, decides which record an update writes; org_id falls back to the query string.
func paymentProfileWriteParams(
	echoCtx *echo.Context, schema *dmodel.ModelSchema, isCreate bool,
) (dmodel.DynamicFields, *ft.ClientErrors) {
	raw := map[string]any{}
	if err := echoCtx.Bind(&raw); err != nil {
		cErrs := ft.ClientErrors{*ft.NewAnonymousValidationError(
			ft.ErrorKey("err_malformed_request"), "malformed request")}
		return nil, &cErrs
	}
	if isCreate {
		if cErrs := unknownProfileFields(raw, schema); cErrs != nil {
			return nil, cErrs
		}
	}

	params := httpserver.FilterToDynamicEntity(raw, schema)
	if config, hasConfig := raw[models.PaymentProfileFieldConfig]; hasConfig {
		params[models.PaymentProfileFieldConfig] = config
	}
	if !isCreate {
		params[basemodel.FieldId] = echoCtx.Param("id")
	}
	if bodyOrgId, _ := params[basemodel.FieldOrgId].(string); bodyOrgId == "" {
		if orgId := echoCtx.QueryParam(basemodel.FieldOrgId); orgId != "" {
			params[basemodel.FieldOrgId] = orgId
		}
	}
	return params, nil
}

// unknownProfileFields reports every body key that names neither a schema field nor "config".
func unknownProfileFields(raw map[string]any, schema *dmodel.ModelSchema) *ft.ClientErrors {
	if schema == nil {
		return nil
	}
	names := make([]string, 0, len(raw))
	for name := range raw {
		if name == models.PaymentProfileFieldConfig {
			continue
		}
		if _, exists := schema.Field(name); !exists {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return nil
	}
	// Map iteration order is random; the response lists the fields deterministically.
	sort.Strings(names)
	cErrs := ft.ClientErrors{}
	for _, name := range names {
		cErrs.Append(*ft.NewValidationError(name, ft.ErrorKey("err_unknown_schema_field"),
			"field is not defined on this schema"))
	}
	return &cErrs
}
