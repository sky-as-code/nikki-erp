package dynamicengines

import (
	"sort"

	"github.com/labstack/echo/v5"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
)

// The payment profile's own binding for create and update.
//
// The engine's built-in binding filters a request body down to the schema fields, which is right
// for every other resource and wrong for this one: "config" is deliberately absent from the schema
// — that is what keeps the credentials out of a column — so the built-in binding would drop it on
// update and reject the request outright on create.
//
// This is the whole of the divergence. Everything after binding is the engine's: the request still
// goes through ExecuteAction, so the permission check, the validation flow and the wrapped main
// process all run exactly as they would on an untouched action.

// paymentProfileWriteHandler serves one of the two write actions.
func paymentProfileWriteHandler(engine drif.DynamicResourceEngine, actionName string) echo.HandlerFunc {
	return func(echoCtx *echo.Context) (err error) {
		defer func() {
			if e := ft.RecoverPanicFailedTo(recover(), "handle REST "+actionName); e != nil {
				err = e
			}
		}()

		reqCtx, err := corectx.AsRequestContext(echoCtx)
		if err != nil {
			return err
		}

		params, cErrs := paymentProfileWriteParams(echoCtx, engine.Schema(), actionName)
		if cErrs != nil {
			return httpserver.JsonBadRequest(echoCtx, *cErrs)
		}

		result, err := engine.ExecuteAction(reqCtx, actionName, params)
		if err != nil {
			return err
		}
		if result.ClientErrors != nil && result.ClientErrors.Count() > 0 {
			return httpserver.JsonBadRequest(echoCtx, result.ClientErrors)
		}
		if !result.HasData {
			return httpserver.JsonBadRequest(echoCtx, ft.ClientErrors{*ft.NewAnonymousNotFoundError()})
		}

		return paymentProfileWriteResponse(echoCtx, actionName, result.Data)
	}
}

// paymentProfileWriteParams binds the body as the engine would, then puts "config" back.
//
// Create rejects a body field that names nothing, matching the built-in create: silently dropping
// it would answer 201 to a request that did not do what the caller asked. Update stays permissive,
// also matching the built-in, because clients round-trip a fetched record back and carry read-only
// keys they never meant to write.
func paymentProfileWriteParams(
	echoCtx *echo.Context, schema *dmodel.ModelSchema, actionName string,
) (dmodel.DynamicFields, *ft.ClientErrors) {
	raw := map[string]any{}
	if err := echoCtx.Bind(&raw); err != nil {
		cErrs := ft.ClientErrors{*ft.NewAnonymousValidationError(
			ft.ErrorKey("err_malformed_request"), "malformed request")}
		return nil, &cErrs
	}

	if actionName == drif.ActionCreate {
		if cErrs := unknownProfileFields(raw, schema); cErrs != nil {
			return nil, cErrs
		}
	}

	params := httpserver.FilterToDynamicEntity(raw, schema)
	if config, hasConfig := raw[models.PaymentProfileFieldConfig]; hasConfig {
		params[models.PaymentProfileFieldConfig] = config
	}
	if actionName == drif.ActionUpdate {
		// The route decides which record is written, never the body.
		params[basemodel.FieldId] = echoCtx.Param("id")
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

// paymentProfileWriteResponse shapes the result exactly as the engine's built-in binding does, so
// a client cannot tell the two write endpoints apart from the rest of the module's CRUD.
func paymentProfileWriteResponse(echoCtx *echo.Context, actionName string, data any) error {
	if actionName == drif.ActionCreate {
		fields, isFields := data.(dmodel.DynamicFields)
		if !isFields {
			return httpserver.JsonCreated(echoCtx, data)
		}
		return httpserver.JsonCreated(echoCtx, httpserver.NewRestCreateResponseDyn(fields))
	}

	mutation, isMutation := data.(dyn.MutateResultData)
	if !isMutation {
		return httpserver.JsonOk(echoCtx, data)
	}

	return httpserver.JsonOk(echoCtx, httpserver.NewRestMutateResponse(mutation))
}
