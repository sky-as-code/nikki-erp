package httpserver

import (
	"github.com/labstack/echo/v4"

	crud "github.com/sky-as-code/nikki-erp/common/crud"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/crud"
)

// BindToDynamicEntity parses the echo request body and returns a DynamicEntity
// containing only the fields defined in the given ModelSchema.
// Minimal type correction is applied via each field's TryConvert; on conversion
// failure the raw parsed value is kept as-is. No validation is performed.
func BindToDynamicEntity(echoCtx echo.Context, entitySchema *dmodel.ModelSchema) (dmodel.DynamicFields, error) {
	var rawBody map[string]any
	if err := echoCtx.Bind(&rawBody); err != nil {
		return nil, err
	}
	return applySchemaFilter(rawBody, entitySchema), nil
}

func applySchemaFilter(rawBody map[string]any, entitySchema *dmodel.ModelSchema) dmodel.DynamicFields {
	result := make(dmodel.DynamicFields)
	for fieldName, field := range entitySchema.Fields() {
		rawVal, exists := rawBody[fieldName]
		if !exists {
			continue
		}
		converted, err := field.DataType().TryConvert(rawVal, field.DataType().Options())
		if err != nil {
			result[fieldName] = rawVal
		} else {
			result[fieldName] = converted
		}
	}
	return result
}

type CmdResult interface {
	GetClientError() *ft.ClientError
	GetHasData() bool
}

// ServeRequestDynamic handles a request using the dynamic entity flow.
// The request body is bound to a DynamicFields map and set on the request object.
// func ServeRequestDynamic[
// 	THttpResp any,
// 	TSvcCommand any,
// 	TSvcResultData dmodel.DynamicModelGetter,
// ](
// 	echoCtx echo.Context,
// 	action string,
// 	createRequestFn func() dmodel.DynamicModelSetter,
// 	serviceFn func(ctx dEnt.Context, cmd TSvcCommand) (*dEnt.OpResult[TSvcResultData], error),
// 	jsonSuccessFn func(echo.Context, any) error,
// ) error {
// 	// TODO: Use `action` for entry and exit logging.
// 	reqCtx := echoCtx.Request().Context().(dEnt.Context)

// 	reqFields := make(dmodel.DynamicFields)
// 	if err := echoCtx.Bind(&reqFields); err != nil {
// 		_, isHttpErr := err.(*echo.HTTPError)
// 		if isHttpErr {
// 			return JsonBadRequest(
// 				echoCtx,
// 				[]any{ft.NewAnonymousValidationError(ft.ErrorKey("err_malformed_request"), "malformed request")},
// 			)
// 		}
// 		return err
// 	}

// 	request := createRequestFn()
// 	request.SetFieldData(reqFields)

// 	cmd, err := modelmapper.CastCopy[*TSvcCommand](request)
// 	if err != nil {
// 		return err
// 	}

// 	result, err := serviceFn(reqCtx, *cmd)
// 	if err != nil {
// 		return err
// 	}

// 	if result.ClientErrors != nil {
// 		return JsonBadRequest(echoCtx, result.ClientErrors)
// 	}

// 	if result.IsEmpty {
// 		cErr := ft.NewAnonymousBusinessViolation(ft.ErrorKey("err_resource_not_found", reqCtx.GetModuleName()), "resource not found")
// 		return JsonBadRequest(echoCtx, []any{cErr})
// 	}

// 	response, err := modelmapper.CastCopy[*THttpResp](map[string]any(result.Data.GetFieldData()))
// 	if err != nil {
// 		return err
// 	}

// 	return jsonSuccessFn(echoCtx, *response)
// }

func ServeRequestDynamic[THttpResp any, TSvcCommand any, TSvcResultData any](
	echoCtx echo.Context,
	serviceFn func(ctx corectx.Context, cmd TSvcCommand) (*crud.OpResult[TSvcResultData], error),
	requestToCommandFn func(requestFields dmodel.DynamicFields) TSvcCommand,
	resultToResponseFn func(data TSvcResultData) THttpResp,
	jsonSuccessFn func(echo.Context, any) error,
) error {
	reqCtx := echoCtx.Request().Context().(corectx.Context)

	reqFields := make(dmodel.DynamicFields)
	if err := echoCtx.Bind(&reqFields); err != nil {
		_, isHttpErr := err.(*echo.HTTPError)
		if isHttpErr {
			return JsonBadRequest(
				echoCtx,
				[]any{ft.NewAnonymousValidationError(ft.ErrorKey("err_malformed_request"), "malformed request")},
			)
		}
		return err
	}

	cmd := requestToCommandFn(reqFields)
	result, err := serviceFn(reqCtx, cmd)

	if err != nil {
		return err
	}

	if result.ClientErrors != nil && result.ClientErrors.Count() > 0 {
		return JsonBadRequest(echoCtx, result.ClientErrors)
	}

	response := resultToResponseFn(result.Data)
	return jsonSuccessFn(echoCtx, response)
}

func ServeRequest2[THttpReq any, THttpResp any, TSvcCommand any, TSvcResultData any](
	echoCtx echo.Context,
	serviceFn func(ctx corectx.Context, cmd TSvcCommand) (*crud.OpResult[TSvcResultData], error),
	requestToCommandFn func(request THttpReq) TSvcCommand,
	resultToResponseFn func(resultData TSvcResultData) THttpResp,
	jsonSuccessFn func(echo.Context, any) error,
) error {
	var request THttpReq
	if err := echoCtx.Bind(&request); err != nil {
		return JsonBadRequest(echoCtx, []any{ft.NewAnonymousValidationError(ft.ErrorKey("err_malformed_request"), "malformed request")})
	}

	cmd := requestToCommandFn(request)
	reqCtx := echoCtx.Request().Context().(corectx.Context)
	result, err := serviceFn(reqCtx, cmd)

	if err != nil {
		return err
	}

	if result.ClientErrors != nil && result.ClientErrors.Count() > 0 {
		return JsonBadRequest(echoCtx, result.ClientErrors)
	}

	response := resultToResponseFn(result.Data)
	return jsonSuccessFn(echoCtx, response)
}

func ServeRequest[THttpReq any, THttpResp any, TSvcCommand any, TSvcResult CmdResult](
	echoCtx echo.Context,
	serviceFn func(ctx corecrud.Context, cmd TSvcCommand) (*TSvcResult, error),
	requestToCommandFn func(request THttpReq) TSvcCommand,
	resultToResponseFn func(result TSvcResult) THttpResp,
	jsonSuccessFn func(echo.Context, any) error,
) error {
	var request THttpReq
	if err := echoCtx.Bind(&request); err != nil {
		return err
	}

	cmd := requestToCommandFn(request)
	reqCtx := echoCtx.Request().Context().(corecrud.Context)
	result, err := serviceFn(reqCtx, cmd)

	if err != nil {
		return err
	}

	if (*result).GetClientError() != nil {
		return JsonBadRequest(echoCtx, (*result).GetClientError())
	}

	if !(*result).GetHasData() {
		cErr := ft.ClientError{
			Code:    "not_found",
			Details: "resource not found",
		}
		return JsonBadRequest(echoCtx, cErr)
	}

	response := resultToResponseFn(*result)
	return jsonSuccessFn(echoCtx, response)
}
