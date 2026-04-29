package middlewares

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/labstack/echo/v5"
	"go.bryk.io/pkg/errors"

	ds "github.com/sky-as-code/nikki-erp/common/datastructure"
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	c "github.com/sky-as-code/nikki-erp/modules/core/httpserver/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/requestguard"
)

// Ensure the request is authorized by any of the processing layers.
func EnsureAuthorized(next echo.HandlerFunc) echo.HandlerFunc {
	return func(echoCtx *echo.Context) error {
		reqCtx, err := corectx.AsRequestContext(echoCtx)
		if err != nil {
			return err
		}

		originalWriter := echoCtx.Response()
		bufferedWriter := httptest.NewRecorder()
		echoCtx.SetResponse(bufferedWriter)
		defer echoCtx.SetResponse(originalWriter)

		nextErr := next(echoCtx)
		if nextErr != nil {
			return nextErr
		}

		echoCtx.SetResponse(originalWriter)

		ctxVal := reqCtx.Value(c.CtxKeyIsAuthorized)
		isAuthorized, ok := ctxVal.(bool)
		if ctxVal == nil || !ok || !isAuthorized {
			return errors.Errorf("No authorization check for endpoint %s", echoCtx.Request().URL.Path)
		}

		for key, values := range bufferedWriter.Header() {
			for _, value := range values {
				originalWriter.Header().Add(key, value)
			}
		}

		statusCode := bufferedWriter.Code
		if statusCode == 0 {
			statusCode = http.StatusOK
		}

		originalWriter.WriteHeader(statusCode)
		_, err = originalWriter.Write(bufferedWriter.Body.Bytes())
		return err
	}
}

func PublicUnauthorized(next echo.HandlerFunc) echo.HandlerFunc {
	return func(echoCtx *echo.Context) error {
		reqCtx, err := corectx.AsRequestContext(echoCtx)
		if err != nil {
			return err
		}
		reqCtx.WithValue(c.CtxKeyIsAuthorized, true)
		return next(echoCtx)
	}
}

// Shortcut for lazy-wrapped SmokeAuthorizeMiddleware.
func SmokeAuthz() echo.MiddlewareFunc {
	return SmokeAuthorizeMiddleware()
}

// Verifies access token and loads user permission to request context.
func SmokeAuthorizeMiddleware() echo.MiddlewareFunc {
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
			reqFingerprint, err := guardSvc.CalcRequestFingerprint(reqCtx, echoCtx.Request())
			if err != nil {
				return err
			}
			// TODO: Check cache
			util.Unused(reqFingerprint)

			resJwt, err := guardSvc.VerifyJwt(reqCtx, echoCtx.Request())
			if err != nil {
				return err
			}
			if !resJwt.IsOk {
				return echoCtx.JSON(http.StatusUnauthorized, resJwt.ClientError)
			}

			reqCtx.WithValue(c.CtxKeyJwtClaims, resJwt.JwtClaims)
			userInfo, err := resJwt.JwtClaims.GetSubject()
			if err != nil {
				return err
			}

			userEmail := strings.Split(userInfo, ":")[0]
			resUser, err := guardSvc.GetUserEntitlements(reqCtx, requestguard.GetUserEntitlementsQuery{
				UserEmail: &userEmail,
			})
			if err != nil {
				return err
			}
			if resUser.ClientErrors.Count() > 0 {
				return errors.Wrap(resUser.ClientErrors.ToError(), "SmokeAuthorizeMiddleware")
			}

			ents := ds.NewSet[string]()
			orgIds := ds.NewSet[model.Id]()
			if resUser.Data.Entitlements != nil {
				ents.AddMany(resUser.Data.Entitlements...)
			}
			if resUser.Data.UserOrgIds != nil {
				orgIds.AddMany(resUser.Data.UserOrgIds...)
			}

			reqCtx.SetUser(resUser.Data.User)
			reqCtx.SetPermissions(corectx.ContextPermissions{
				IsOwner:      resUser.Data.IsOwner,
				Entitlements: ents,
				UserId:       resUser.Data.UserId,
				UserOrgIds:   orgIds,
				OrgUnitId:    resUser.Data.OrgUnitId,
				OrgUnitOrgId: resUser.Data.OrgUnitOrgId,
			})

			return next(echoCtx)
		}
	}
}
