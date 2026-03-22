package httpserver

import (
	"github.com/labstack/echo/v4"

	"github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/modelmapper"
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

func ServeRequestDynamic[
	THttpResp any,
	TSvcCommand any,
	TSvcResultData schema.DynamicModelGetter,
](
	echoCtx echo.Context,
	action string,
	createRequestFn func() schema.DynamicModelSetter,
	serviceFn func(ctx dEnt.Context, cmd TSvcCommand) (*dEnt.OpResult[TSvcResultData], error),
	jsonSuccessFn func(echo.Context, any) error,
) error {
	// TODO: Use `action` for entry and exit logging.

	reqFields := make(schema.DynamicFields)
	err := echoCtx.Bind(&reqFields)
	if err != nil {
		return err
	}

	request := createRequestFn()
	request.SetFieldData(reqFields)

	reqCtx := echoCtx.Request().Context().(dEnt.Context)
	cmd, err := modelmapper.CastCopy[*TSvcCommand](request)
	if err != nil {
		return err
	}

	result, err := serviceFn(reqCtx, *cmd)
	if err != nil {
		return err
	}

	if result.ClientErrors != nil {
		return JsonBadRequest(echoCtx, result.ClientErrors)
	}

	if result.IsEmpty {
		cErr := ft.ClientError{
			Code:    "not_found",
			Details: "resource not found",
		}
		return JsonBadRequest(echoCtx, cErr)
	}

	response, err := modelmapper.MapToStruct[*THttpResp](result.Data.GetFieldData())
	if err != nil {
		return err
	}
	return jsonSuccessFn(echoCtx, *response)
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
