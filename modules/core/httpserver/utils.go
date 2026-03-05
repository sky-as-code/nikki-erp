package httpserver

import (
	stdErr "errors"
	"net/http"

	"github.com/labstack/echo/v4"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

type CmdResult interface {
	GetClientError() *ft.ClientError
	GetHasData() bool
}

func HandleBindError(echoCtx echo.Context, err error) error {
	return JsonBadRequest(echoCtx, &ft.ClientError{
		Code:    "bad_request",
		Details: ft.ValidationErrors{"request": err.Error()},
	})
}

// HandleServiceError maps a service-layer error to a JSON response.
// ClientError is returned with the appropriate HTTP status.
// Any other error returns a generic 500 to avoid leaking internal details.
func HandleServiceError(echoCtx echo.Context, err error) error {
	var clientErr *ft.ClientError
	if stdErr.As(err, &clientErr) {
		status := mapClientErrorStatus(clientErr.Code)
		return echoCtx.JSON(status, clientErr)
	}

	logging.Logger().Error("unexpected service error", err)
	return echoCtx.JSON(http.StatusInternalServerError, &ft.ClientError{
		Code:    "internal_error",
		Details: "an unexpected error occurred",
	})
}

// HandleResultError inspects a CmdResult and returns the appropriate JSON error.
// Returns nil when the result is successful so the caller can build the response.
func HandleResultError(echoCtx echo.Context, result CmdResult) error {
	if (result).GetClientError() != nil {
		return JsonBadRequest(echoCtx, (result).GetClientError())
	}

	if !(result).GetHasData() {
		cErr := ft.ClientError{
			Code:    "not_found",
			Details: "resource not found",
		}
		return JsonBadRequest(echoCtx, cErr)
	}

	return nil
}

func mapClientErrorStatus(code string) int {
	switch code {
	case "duplicate_name":
		return http.StatusConflict
	default:
		return http.StatusBadRequest
	}
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
