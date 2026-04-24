package httpserver

import (
	"errors"
	"mime/multipart"
	"reflect"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/crud"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver/middlewares"
)

// BindToDynamicEntity parses the echo request body and returns a DynamicEntity
// containing only the fields defined in the given ModelSchema.
// Minimal type correction is applied via each field's TryConvert; on conversion
// failure the raw parsed value is kept as-is. No validation is performed.
func BindToDynamicEntity(echoCtx *echo.Context, entitySchema *dmodel.ModelSchema) (dmodel.DynamicFields, error) {
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

func BindFormFile(echoCtx *echo.Context, cmd any) error {
	requestType := reflect.TypeOf(cmd).Elem()
	requestValue := reflect.ValueOf(cmd).Elem()
	form, err := echoCtx.MultipartForm()
	if err != nil {
		return err
	}

	sliceFileType := reflect.TypeFor[[]*multipart.FileHeader]()
	fileType := reflect.TypeFor[*multipart.FileHeader]()

	for i := range requestType.NumField() {
		fieldType := requestType.Field(i)
		fieldValue := requestValue.Field(i)
		fileTagVal := fieldType.Tag.Get("form-file")

		if !fieldValue.CanSet() || fileTagVal == "" {
			continue
		}

		if fieldType.Type.Kind() == reflect.Slice {
			files := form.File[fileTagVal]
			if len(files) > 0 {
				if sliceFileType.AssignableTo(fieldType.Type) {
					fieldValue.Set(reflect.ValueOf(files))
				}
			}

		} else {
			file, err := echoCtx.FormFile(fileTagVal)
			if err != nil {
				continue
			}

			if fileType.AssignableTo(fieldType.Type) {
				fieldValue.Set(reflect.ValueOf(file))
			}
		}
	}

	return nil
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
// 	echoCtx *echo.Context,
// 	action string,
// 	createRequestFn func() dmodel.DynamicModelSetter,
// 	serviceFn func(ctx dEnt.Context, cmd TSvcCommand) (*dEnt.OpResult[TSvcResultData], error),
// 	jsonSuccessFn func(*echo.Context, any) error,
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

// 	if !result.HasData {
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
	echoCtx *echo.Context,
	serviceFn func(ctx corectx.Context, cmd TSvcCommand) (*dyn.OpResult[TSvcResultData], error),
	requestToCommandFn func(requestFields dmodel.DynamicFields) TSvcCommand,
	resultToResponseFn func(data TSvcResultData) THttpResp,
	jsonSuccessFn func(*echo.Context, any) error,
) error {
	reqCtx := echoCtx.Request().Context().(corectx.Context)

	reqFields := make(map[string]any)
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

	if !result.HasData {
		cErrs := ft.ClientErrors{*ft.NewAnonymousNotFoundError()}
		return JsonBadRequest(echoCtx, cErrs)
	}

	response := resultToResponseFn(result.Data)
	return jsonSuccessFn(echoCtx, response)
}

func ServeRequestFormData[TBinding any, THttpResp any, TSvcCommand, TSvcResultData any](
	echoCtx *echo.Context,
	serviceFn func(ctx corectx.Context, cmd TSvcCommand) (*dyn.OpResult[TSvcResultData], error),
	requestToCommandFn func(request TBinding) TSvcCommand,
	resultToResponseFn func(resultData TSvcResultData) THttpResp,
	jsonSuccessFn func(*echo.Context, any) error,
	skipNotFoundError ...bool,
) error {
	var request TBinding

	requestType := reflect.TypeOf(request)
	if requestType.Kind() != reflect.Struct {
		panic("TBinding must be a struct")
	}

	err := echoCtx.Bind(&request)
	if err != nil {
		return err
	}

	err = BindFormFile(echoCtx, &request)
	if err != nil {
		return err
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

	if !result.HasData && (len(skipNotFoundError) == 0 || !skipNotFoundError[0]) {
		cErrs := ft.ClientErrors{*ft.NewAnonymousNotFoundError()}
		return JsonBadRequest(echoCtx, cErrs)
	}

	response := resultToResponseFn(result.Data)
	return jsonSuccessFn(echoCtx, response)
}

func ServeRequest2[THttpReq any, THttpResp any, TSvcCommand any, TSvcResultData any](
	echoCtx *echo.Context,
	serviceFn func(ctx corectx.Context, cmd TSvcCommand) (*dyn.OpResult[TSvcResultData], error),
	requestToCommandFn func(request THttpReq) TSvcCommand,
	resultToResponseFn func(resultData TSvcResultData) THttpResp,
	jsonSuccessFn func(*echo.Context, any) error,
	skipNotFoundError ...bool,
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

	if !result.HasData && (len(skipNotFoundError) == 0 || !skipNotFoundError[0]) {
		cErrs := ft.ClientErrors{*ft.NewAnonymousNotFoundError()}
		return JsonBadRequest(echoCtx, cErrs)
	}

	response := resultToResponseFn(result.Data)
	return jsonSuccessFn(echoCtx, response)
}

func ServeRequest[THttpReq any, THttpResp any, TSvcCommand any, TSvcResult CmdResult](
	echoCtx *echo.Context,
	serviceFn func(ctx corecrud.Context, cmd TSvcCommand) (*TSvcResult, error),
	requestToCommandFn func(request THttpReq) TSvcCommand,
	resultToResponseFn func(result TSvcResult) THttpResp,
	jsonSuccessFn func(*echo.Context, any) error,
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

func ItsMeMario[T any](me T) T {
	return me
}

func ServeCreate[
	TSvcCommand any,
	TSvcCommandPtr dyn.DynamicModelSetterPtr[TSvcCommand],
	TSvcResultData dmodel.DynamicModelGetter,
](
	action string,
	echoCtx *echo.Context,
	cmd TSvcCommandPtr,
	serviceFn func(ctx corectx.Context, cmd TSvcCommand) (*dyn.OpResult[TSvcResultData], error),
) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST "+action); e != nil {
			err = e
		}
	}()
	return ServeRequestDynamic(
		echoCtx,
		serviceFn,
		func(requestFields dmodel.DynamicFields) TSvcCommand {
			cmd.SetFieldData(requestFields)
			return *cmd
		},
		func(data TSvcResultData) RestCreateResponse {
			return *NewRestCreateResponseDyn(data.GetFieldData())
		},
		JsonCreated,
	)
}

func ServeExists[
	TSvcCommand dyn.ExistsQueryShape,
](
	action string,
	echoCtx *echo.Context,
	serviceFn func(ctx corectx.Context, cmd TSvcCommand) (*dyn.OpResult[dyn.ExistsResultData], error),
) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST "+action); e != nil {
			err = e
		}
	}()
	err = ServeRequest2(
		echoCtx,
		serviceFn,
		ItsMeMario,
		ItsMeMario,
		JsonOk,
	)
	return err
}

func ServeGetOne[
	TSvcQuery any,
	TDomain dmodel.DynamicModelGetter,
](
	action string,
	echoCtx *echo.Context,
	serviceFn func(ctx corectx.Context, cmd TSvcQuery) (*dyn.OpResult[TDomain], error),
) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST "+action); e != nil {
			err = e
		}
	}()
	err = ServeRequest2(
		echoCtx,
		serviceFn,
		ItsMeMario,
		func(data TDomain) dmodel.DynamicFields {
			return data.GetFieldData()
		},
		JsonOk,
	)
	return err
}

// Use this function for mutation operations (delete, manageM2m, setIsArchived, etc.),
// but not for update.
func ServeGeneralMutate[
	TSvcCommand any,
](
	action string,
	echoCtx *echo.Context,
	serviceFn func(ctx corectx.Context, cmd TSvcCommand) (*dyn.OpResult[dyn.MutateResultData], error),
) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST "+action); e != nil {
			err = e
		}
	}()
	err = ServeRequest2(
		echoCtx,
		serviceFn,
		ItsMeMario,
		NewRestMutateResponse,
		JsonOk,
	)
	return err
}

func ServeSearch[
	TSvcQuery any,
	TDomain dmodel.DynamicModelGetter,
](
	action string,
	echoCtx *echo.Context,
	serviceFn func(ctx corectx.Context, cmd TSvcQuery) (*dyn.OpResult[dyn.PagedResultData[TDomain]], error),
) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST "+action); e != nil {
			err = e
		}
	}()
	err = ServeRequest2(
		echoCtx,
		serviceFn,
		ItsMeMario,
		NewSearchResponseDyn,
		JsonOk,
		true,
	)
	return err
}

func ServeUpdate[
	TSvcCommand any,
	TSvcCommandPtr dyn.DynamicModelSetterPtr[TSvcCommand],
](
	action string,
	echoCtx *echo.Context,
	cmd TSvcCommandPtr,
	serviceFn func(ctx corectx.Context, cmd TSvcCommand) (*dyn.OpResult[dyn.MutateResultData], error),
) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST "+action); e != nil {
			err = e
		}
	}()
	return ServeRequestDynamic(
		echoCtx,
		serviceFn,
		func(requestFields dmodel.DynamicFields) TSvcCommand {
			cmd.SetFieldData(requestFields)
			return *cmd
		},
		func(data dyn.MutateResultData) RestMutateResponse {
			return NewRestMutateResponse(data)
		},
		JsonOk,
	)
}

func GetUserEmailFromContext(ctx corectx.Context) (string, error) {
	claims, ok := ctx.Value(middlewares.CtxKeyJwtClaims).(jwt.Claims)
	if !ok {
		return "", errors.New("User not login")
	}

	userInfo, err := claims.GetSubject()
	if err != nil {
		return "", err
	}
	userEmail := strings.Split(userInfo, ":")[0]
	return userEmail, nil
}
