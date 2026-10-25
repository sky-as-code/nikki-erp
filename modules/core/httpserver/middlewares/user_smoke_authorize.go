package middlewares

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
	"go.bryk.io/pkg/errors"

	ds "github.com/sky-as-code/nikki-erp/common/datastructure"
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
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

		ctxVal := reqCtx.Value(c.CtxKeyUserIsAuthorized)
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
		reqCtx.WithValue(c.CtxKeyUserIsAuthorized, true)
		return next(echoCtx)
	}
}

// Shortcut for lazy-wrapped UserSmokeAuthorizeMiddleware.
func UserSmokeAuthz() echo.MiddlewareFunc {
	return UserSmokeAuthorizeMiddleware()
}

// Verifies the user access token and loads the user's permissions into the request context.
func UserSmokeAuthorizeMiddleware() echo.MiddlewareFunc {
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

			resJwt, err := guardSvc.VerifyUserAccessToken(reqCtx, echoCtx.Request())
			if err != nil {
				return err
			}
			if !resJwt.IsOk {
				return echoCtx.JSON(http.StatusUnauthorized, ft.ClientErrors{*resJwt.ClientError})
			}
			// With CORE.REQUEST_GUARD.ACCESS_TOKEN_ENABLED=false there are no claims, hence no user:
			// the request continues anonymously and permission checks downstream decide.
			if resJwt.JwtClaims == nil {
				return next(echoCtx)
			}

			reqCtx.WithValue(c.CtxKeyJwtClaims, resJwt.JwtClaims)
			err = loadUserPermissions(reqCtx, guardSvc, resJwt.JwtClaims)
			if err != nil {
				return err
			}
			return next(echoCtx)
		}
	}
}

func loadUserPermissions(
	reqCtx corectx.Context, guardSvc requestguard.RequestGuardService, claims jwt.Claims,
) error {
	userInfo, err := claims.GetSubject()
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
		return errors.Wrap(resUser.ClientErrors.ToError(), "UserSmokeAuthorizeMiddleware")
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
		// This middleware authenticates a person through a JWT, so the actor is always a
		// user. A service reaches the application by another route and sets its own
		// principal there.
		Principal: corectx.Principal{
			Kind:        corectx.PrincipalKindUser,
			Id:          resUser.Data.UserId,
			DisplayName: userEmail,
		},
		UserOrgIds:   orgIds,
		OrgUnitId:    resUser.Data.OrgUnitId,
		OrgUnitOrgId: resUser.Data.OrgUnitOrgId,
	})
	return nil
}
