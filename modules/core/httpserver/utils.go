package httpserver

import (
	"github.com/labstack/echo/v4"

	"github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	dEnt "github.com/sky-as-code/nikki-erp/modules/core/dynamicentity"
)

// BindToDynamicEntity parses the echo request body and returns a DynamicEntity
// containing only the fields defined in the given EntitySchema.
// Minimal type correction is applied via each field's TryConvert; on conversion
// failure the raw parsed value is kept as-is. No validation is performed.
func BindToDynamicEntity(echoCtx echo.Context, entitySchema *schema.EntitySchema) (schema.DynamicFields, error) {
	var rawBody map[string]any
	if err := echoCtx.Bind(&rawBody); err != nil {
		return nil, err
	}
	return applySchemaFilter(rawBody, entitySchema), nil
}

func applySchemaFilter(rawBody map[string]any, entitySchema *schema.EntitySchema) schema.DynamicFields {
	result := make(schema.DynamicFields)
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
// 	TSvcResultData schema.DynamicModelGetter,
// ](
// 	echoCtx echo.Context,
// 	action string,
// 	createRequestFn func() schema.DynamicModelSetter,
// 	serviceFn func(ctx dEnt.Context, cmd TSvcCommand) (*dEnt.OpResult[TSvcResultData], error),
// 	jsonSuccessFn func(echo.Context, any) error,
// ) error {
// 	// TODO: Use `action` for entry and exit logging.
// 	reqCtx := echoCtx.Request().Context().(dEnt.Context)

// 	reqFields := make(schema.DynamicFields)
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
	serviceFn func(ctx dEnt.Context, cmd TSvcCommand) (*dEnt.OpResult[TSvcResultData], error),
	requestToCommandFn func(requestFields schema.DynamicFields) TSvcCommand,
	resultToResponseFn func(data TSvcResultData) THttpResp,
	jsonSuccessFn func(echo.Context, any) error,
) error {
	reqCtx := echoCtx.Request().Context().(dEnt.Context)

	reqFields := make(schema.DynamicFields)
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
	serviceFn func(ctx dEnt.Context, cmd TSvcCommand) (*dEnt.OpResult[TSvcResultData], error),
	requestToCommandFn func(request THttpReq) TSvcCommand,
	resultToResponseFn func(resultData TSvcResultData) THttpResp,
	jsonSuccessFn func(echo.Context, any) error,
) error {
	var request THttpReq
	if err := echoCtx.Bind(&request); err != nil {
		return err
	}

	cmd := requestToCommandFn(request)
	reqCtx := echoCtx.Request().Context().(dEnt.Context)
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
	serviceFn func(ctx crud.Context, cmd TSvcCommand) (*TSvcResult, error),
	requestToCommandFn func(request THttpReq) TSvcCommand,
	resultToResponseFn func(result TSvcResult) THttpResp,
	jsonSuccessFn func(echo.Context, any) error,
) error {
	var request THttpReq
	if err := echoCtx.Bind(&request); err != nil {
		return err
	}

	cmd := requestToCommandFn(request)
	reqCtx := echoCtx.Request().Context().(crud.Context)
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
