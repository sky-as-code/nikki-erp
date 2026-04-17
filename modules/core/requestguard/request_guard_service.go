package requestguard

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"go.bryk.io/pkg/errors"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	coretoken "github.com/sky-as-code/nikki-erp/modules/core/authtoken"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type StaticRequestGuardServiceParams struct {
	dig.In

	ConfigSvc config.ConfigService
	TokenSvc  coretoken.AuthTokenService
}

func NewStaticRequestGuardServiceImpl(params StaticRequestGuardServiceParams) RequestGuardService {
	return &StaticRequestGuardServiceImpl{
		configSvc: params.ConfigSvc,
		tokenSvc:  params.TokenSvc,
	}
}

type StaticRequestGuardServiceImpl struct {
	configSvc      config.ConfigService
	tokenSvc       coretoken.AuthTokenService
	corsMiddleware echo.MiddlewareFunc
}

func (this *StaticRequestGuardServiceImpl) CalcRequestFingerprint(_ corectx.Context, request *http.Request) (fingerprint string, err error) {
	if this.configSvc.GetBool(c.RequestGuardAccessTokenEnabled) {
		rawToken := this.bearerAccessToken(request)
		parts := strings.Split(rawToken, ".")
		if rawToken == "" || len(parts) != 3 {
			return "", nil
		}
		return parts[0] + "." + parts[1], nil
	}
	return "", errors.New("not implemented")
}

func (this *StaticRequestGuardServiceImpl) GetCorsMiddleware(_ corectx.Context) (echo.MiddlewareFunc, error) {
	if !this.configSvc.GetBool(c.HttpCorsEnabled) {
		return nil, nil
	}

	if this.corsMiddleware == nil {
		this.corsMiddleware = middleware.CORSWithConfig(this.configCors())
	}
	return this.corsMiddleware, nil
}

func (this *StaticRequestGuardServiceImpl) configCors() middleware.CORSConfig {
	corsOrigins := this.configSvc.GetStrArr(c.HttpCorsOrigins)
	corsHeaders := this.configSvc.GetStrArr(c.HttpCorsHeaders)
	corsMethods := this.configSvc.GetStrArr(c.HttpCorsMethods)

	return middleware.CORSConfig{
		AllowOrigins: corsOrigins,
		AllowHeaders: corsHeaders,
		AllowMethods: corsMethods,
	}
}

func (this *StaticRequestGuardServiceImpl) VerifyTrustedConnection(ctx corectx.Context, request *http.Request) (result *VerifyRequestResult, err error) {
	return &VerifyRequestResult{
		IsOk: true,
	}, nil
}

func (this *StaticRequestGuardServiceImpl) VerifyJwt(ctx corectx.Context, request *http.Request) (*VerifyRequestResult, error) {
	cfg := this.configSvc
	if !cfg.GetBool(c.RequestGuardAccessTokenEnabled) {
		return &VerifyRequestResult{
			IsOk: true,
		}, nil
	}

	rawToken := this.bearerAccessToken(request)
	if rawToken == "" {
		return jwtInvalidFailure(), nil
	}

	verifyResult, err := this.tokenSvc.VerifyJwt(ctx, coretoken.VerifyJwtParam{
		Token: rawToken,
	})
	if err != nil {
		return jwtMalformedFailure(), nil
	}
	if !verifyResult.IsOk {
		return jwtInvalidFailure(), nil
	}

	if this.configSvc.GetBool(c.RequestGuardAccessTokenDpopEnabled) {
		result, dpopErr := this.VerifyJwtDpop(ctx, request)
		if result != nil || dpopErr != nil {
			return result, dpopErr
		}
	}
	return &VerifyRequestResult{
		IsOk:      true,
		JwtClaims: verifyResult.Claims,
	}, nil
}

func (this *StaticRequestGuardServiceImpl) bearerAccessToken(request *http.Request) string {
	headerName := strings.TrimSpace(this.configSvc.GetStr(c.RequestGuardAccessTokenHttpHeaderName))
	prefix := strings.TrimSpace(this.configSvc.GetStr(c.RequestGuardAccessTokenHttpHeaderPrefix))
	auth := strings.TrimSpace(request.Header.Get(headerName))

	if auth == "" {
		return ""
	}
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return ""
	}
	raw := strings.TrimSpace(auth[len(prefix):])
	if raw == "" {
		return ""
	}
	return raw
}

func jwtInvalidFailure() *VerifyRequestResult {
	return &VerifyRequestResult{
		IsOk: false,
		ClientError: ft.NewAuthorizationError(
			ft.ErrorKey("err_invalid_access_token", "authorize"),
			"Invalid or expired access token.",
		),
	}
}

func jwtMalformedFailure() *VerifyRequestResult {
	return &VerifyRequestResult{
		IsOk: false,
		ClientError: ft.NewAuthorizationError(
			ft.ErrorKey("err_malformed_access_token", "authorize"),
			"Malformed access token.",
		),
	}
}

// Verify JWT DPoP (OAuth2 Demonstraing Proof of Possession)
func (this *StaticRequestGuardServiceImpl) VerifyJwtDpop(ctx corectx.Context, request *http.Request) (*VerifyRequestResult, error) {

	return nil, nil
}

func (this *StaticRequestGuardServiceImpl) VerifySessionBlacklist(ctx corectx.Context, request *http.Request) (*VerifyRequestResult, error) {

	return nil, nil
}
