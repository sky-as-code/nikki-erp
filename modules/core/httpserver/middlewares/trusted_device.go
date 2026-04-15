package middlewares

import (
	"net/http"

	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/requestguard"
)

func TrustedDeviceMiddleware() echo.MiddlewareFunc {
	var guardSvc requestguard.RequestGuardService
	deps.Invoke(func(guard requestguard.RequestGuardService) {
		guardSvc = guard
	})
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(echoCtx *echo.Context) error {
			reqCtx, err := corectx.AsRequestContext(echoCtx)
			if err != nil {
				return err
			}
			guardResult, err := guardSvc.VerifyTrustedConnection(reqCtx, echoCtx.Request())
			if err != nil {
				return err
			}
			if !guardResult.IsOk {
				return echoCtx.JSON(http.StatusUnauthorized, guardResult.ClientError)
			}

			return next(echoCtx)
		}
	}
}
