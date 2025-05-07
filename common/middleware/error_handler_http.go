package middleware

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sky-as-code/nikki-erp/common/env"
	"github.com/sky-as-code/nikki-erp/common/fault"
)

func CustomHttpErrorHandler(defaultHandler echo.HTTPErrorHandler) echo.HTTPErrorHandler {
	return func(err error, ctx echo.Context) {
		if ctx.Response().Committed {
			// Response already sent
			return
		}
		// transErr := transformError(err)
		// handleHttpError(transErr, ctx)
		if !handleerror(err, ctx) {
			defaultHandler(err, ctx)
		}
	}
}

/*
func transformError(err error) error {
	_, isEchoHttpErr := err.(*echo.HTTPError)
	// _echoHttpErr, isEchoHttpErr := err.(*echo.HTTPError)

	if isEchoHttpErr {
		return err //fault.NewBusinessError(handler.InternalFailure, echoHttpErr.Message.(string))
	}
	return err
}

func handleHttpError(err error, ctx echo.Context) {
	httpErr, isHttpErr := err.(fault.HttpError)
	logger := ctx.Logger()

	if isHttpErr {
		logger.Warnf("HttpError: %s", err.Error())
		fault.PanicOnErr(ctx.JSON(httpErr.StatusCode(), httpErr.Error()))
		return
	}
	if handleerror(err, ctx) {
		return
	}
	logger.Errorf("UnknownError: %s", err.Error())
	ctx.Response().WriteHeader(http.StatusInternalServerError)
}
*/

func handleerror(err error, ctx echo.Context) bool {
	var bizErr fault.BusinessError
	var techErr fault.TechnicalError
	var valErr fault.ValidationError
	logger := ctx.Logger()

	if errors.As(err, &valErr) {
		logger.Warnf("ValidationError: %s", err.Error())
		clientHttpErr := fault.WrapClientHttpError(valErr)
		fault.PanicOnErr(ctx.JSON(http.StatusUnprocessableEntity, clientHttpErr))
		return true
	}
	if errors.As(err, &bizErr) {
		logger.Warnf("BusinessError: %s", err.Error())
		clientHttpErr := fault.WrapClientHttpError(bizErr)
		fault.PanicOnErr(ctx.JSON(http.StatusUnprocessableEntity, clientHttpErr))
		return true
	}
	if errors.As(err, &techErr) && env.IsLocal() {
		logger.Errorf("TechnicalError: %s", err.Error())
		serverHttpErr := fault.WrapInternalServerHttpError(techErr)
		fault.PanicOnErr(ctx.JSON(http.StatusInternalServerError, serverHttpErr))
		return true
	}
	return false
}
