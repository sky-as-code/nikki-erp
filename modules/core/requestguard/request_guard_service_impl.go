package requestguard

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"go.bryk.io/pkg/errors"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	coretoken "github.com/sky-as-code/nikki-erp/modules/core/authtoken"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

type StaticRequestGuardServiceParams struct {
	dig.In

	ConfigSvc config.ConfigService
	CqrsBus   cqrs.CqrsBus
	TokenSvc  coretoken.AuthTokenService
}

func NewStaticRequestGuardServiceImpl(params StaticRequestGuardServiceParams) RequestGuardService {
	return &StaticRequestGuardServiceImpl{
		configSvc: params.ConfigSvc,
		cqrsBus:   params.CqrsBus,
		tokenSvc:  params.TokenSvc,
	}
}

type StaticRequestGuardServiceImpl struct {
	corsMiddleware echo.MiddlewareFunc
	configSvc      config.ConfigService
	cqrsBus        cqrs.CqrsBus
	tokenSvc       coretoken.AuthTokenService
}

func (this *StaticRequestGuardServiceImpl) CalcRequestFingerprint(_ corectx.Context, request *http.Request) (fingerprint string, err error) {
	if this.configSvc.GetBool(c.RequestGuardAccessTokenEnabled) {
		rawToken := this.userAccessToken(request)
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

func (this *StaticRequestGuardServiceImpl) GetUserEntitlements(
	ctx corectx.Context, query GetUserEntitlementsQuery,
) (*GetUserEntitlementsResult, error) {
	extQuery := ExtGetUserEntitlementsQuery{
		UserEmail: query.UserEmail,
	}
	extResult := ExtGetUserEntitlementsResult{}
	err := this.cqrsBus.Request(ctx, &extQuery, &extResult)
	if err != nil {
		return nil, err
	}
	if extResult.ClientErrors.Count() > 0 {
		return nil, errors.Wrap(extResult.ClientErrors.ToError(), "PermissionLoaderImpl.GetUserEntitlements")
	}
	if !extResult.HasData {
		return &GetUserEntitlementsResult{
			HasData: false,
		}, nil
	}

	return &GetUserEntitlementsResult{
		Data:    extResult.Data,
		HasData: extResult.HasData,
	}, nil
}

func (this *StaticRequestGuardServiceImpl) VerifyUserAccessToken(
	ctx corectx.Context, request *http.Request,
) (*VerifyRequestResult, error) {
	if !this.configSvc.GetBool(c.RequestGuardAccessTokenEnabled) {
		return &VerifyRequestResult{IsOk: true}, nil
	}

	result, claims, err := this.verifyRawToken(ctx, this.userAccessToken(request))
	if result != nil || err != nil {
		return result, err
	}

	if this.configSvc.GetBool(c.RequestGuardAccessTokenDpopEnabled) {
		result, dpopErr := this.VerifyJwtDpop(ctx, request)
		if result != nil || dpopErr != nil {
			return result, dpopErr
		}
	}
	return &VerifyRequestResult{
		IsOk:      true,
		JwtClaims: claims,
	}, nil
}

// verifyRawToken returns a non-nil failure result when the token is absent, malformed or invalid,
// otherwise the verified claims.
func (this *StaticRequestGuardServiceImpl) verifyRawToken(
	ctx corectx.Context, rawToken string,
) (*VerifyRequestResult, jwt.MapClaims, error) {
	if rawToken == "" {
		return jwtInvalidFailure(), nil, nil
	}
	verifyResult, err := this.tokenSvc.VerifyJwt(ctx, coretoken.VerifyJwtParam{
		Token: rawToken,
	})
	if err != nil {
		return jwtMalformedFailure(), nil, nil
	}
	if !verifyResult.IsOk {
		return jwtInvalidFailure(), nil, nil
	}
	return nil, verifyResult.Claims, nil
}

func (this *StaticRequestGuardServiceImpl) userAccessToken(request *http.Request) string {
	headerName := strings.TrimSpace(this.configSvc.GetStr(c.RequestGuardAccessTokenHttpHeaderName))
	prefix := this.configSvc.GetStr(c.RequestGuardAccessTokenHttpHeaderPrefix, "")
	return extractPrefixedToken(request.Header.Get(headerName), prefix)
}

// extractPrefixedToken reads "{prefix} {token}" from a header value. An empty prefix means the
// header carries the raw token.
func extractPrefixedToken(headerValue string, prefix string) string {
	value := strings.TrimSpace(headerValue)
	prefix = strings.TrimSpace(prefix)
	if value == "" {
		return ""
	}
	if prefix == "" {
		return value
	}
	if len(value) <= len(prefix) || !strings.EqualFold(value[:len(prefix)], prefix) {
		return ""
	}
	if separator := value[len(prefix)]; separator != ' ' && separator != '\t' {
		return ""
	}
	return strings.TrimSpace(value[len(prefix):])
}

func jwtInvalidFailure() *VerifyRequestResult {
	return &VerifyRequestResult{
		IsOk: false,
		ClientError: ft.NewAuthorizationError(
			ft.ErrorKey("err_invalid_access_token"),
			"Invalid or expired access token.",
		),
	}
}

func jwtMalformedFailure() *VerifyRequestResult {
	return &VerifyRequestResult{
		IsOk: false,
		ClientError: ft.NewAuthorizationError(
			ft.ErrorKey("err_malformed_access_token"),
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
