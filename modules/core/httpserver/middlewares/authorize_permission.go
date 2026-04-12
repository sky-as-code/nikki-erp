package middlewares

import (
	"net/http"

	"github.com/labstack/echo/v4"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	ext "github.com/sky-as-code/nikki-erp/modules/core/httpserver/external"
	"github.com/sky-as-code/nikki-erp/modules/core/requestguard"
)

// Short-hand for AuthorizePermissionMiddleware
func Authorized(actionCode, resourceCode, scope string) echo.MiddlewareFunc {
	return AuthorizePermissionMiddleware(AuthzPermMiddlewareParams{
		ActionCode:   actionCode,
		ResourceCode: resourceCode,
		Scope:        scope,
	})
}

type AuthzPermMiddlewareParams struct {
	ActionCode   string
	ResourceCode string
	Scope        string
}

func AuthorizePermissionMiddleware(params AuthzPermMiddlewareParams) echo.MiddlewareFunc {
	var permissionSvc ext.PermissionExtService
	var guardSvc requestguard.RequestGuardService
	deps.Invoke(func(permission ext.PermissionExtService, guard requestguard.RequestGuardService) {
		permissionSvc = permission
		guardSvc = guard
	})
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(echoCtx echo.Context) error {
			reqCtx, err := corectx.AsRequestContext(echoCtx)
			if err != nil {
				return err
			}

			reqFingerprint, err := guardSvc.CalcRequestFingerprint(reqCtx, echoCtx.Request())
			if err != nil {
				return err
			}
			// TODO: Check cache
			util.Unused(reqFingerprint)

			result, err := guardSvc.VerifyJwt(reqCtx, echoCtx.Request())
			if err != nil {
				return err
			}
			if !result.IsOk {
				return echo.NewHTTPError(result.HttpStatus, result.ClientError)
			}

			reqCtx.WithValue(CtxKeyJwtClaims, result.JwtClaims)
			userEmail, err := result.JwtClaims.GetSubject()
			if err != nil {
				return err
			}

			isAuthorized, err := permissionSvc.IsAuthorized(reqCtx, ext.IsAuthorizedQuery{
				UserEmail:    &userEmail,
				ActionCode:   params.ActionCode,
				ResourceCode: params.ResourceCode,
				Scope:        params.Scope,
			})
			if err != nil {
				return err
			}
			if !isAuthorized {
				return echo.NewHTTPError(http.StatusForbidden, ft.ClientErrorItem{
					Key:     ft.ErrorKey("not_authorized", "authorize"),
					Message: "not authorized",
				})
			}

			return next(echoCtx)
		}
	}
}
