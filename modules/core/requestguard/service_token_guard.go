package requestguard

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

const (
	defaultServiceTokenHeaderName   = "Authorization"
	defaultServiceTokenHeaderPrefix = "Bearer"
)

func (this *StaticRequestGuardServiceImpl) VerifyServiceAccessToken(
	ctx corectx.Context, request *http.Request, allowedScopes []string,
) (*VerifyRequestResult, error) {
	if !this.configSvc.GetBool(c.RequestGuardAccessTokenEnabled) {
		return &VerifyRequestResult{IsOk: true}, nil
	}

	result, claims, err := this.verifyRawToken(ctx, this.serviceAccessToken(request))
	if result != nil || err != nil {
		return result, err
	}
	if !scopesIntersect(claimScopes(claims), allowedScopes) {
		return scopeDeniedFailure(), nil
	}
	return &VerifyRequestResult{
		IsOk:      true,
		JwtClaims: claims,
	}, nil
}

func (this *StaticRequestGuardServiceImpl) serviceAccessToken(request *http.Request) string {
	headerName := strings.TrimSpace(this.configSvc.GetStr(
		c.RequestGuardServiceTokenHttpHeaderName, defaultServiceTokenHeaderName))
	prefix := this.configSvc.GetStr(c.RequestGuardServiceTokenHttpHeaderPrefix, defaultServiceTokenHeaderPrefix)
	return extractPrefixedToken(request.Header.Get(headerName), prefix)
}

// claimScopes reads the "scope" claim, which RFC 8693 defines as a space-separated string; some
// issuers emit an array instead, so both are accepted.
func claimScopes(claims jwt.MapClaims) []string {
	switch raw := claims["scope"].(type) {
	case string:
		return strings.Fields(raw)
	case []string:
		return raw
	case []any:
		scopes := make([]string, 0, len(raw))
		for _, item := range raw {
			if str, ok := item.(string); ok {
				scopes = append(scopes, strings.Fields(str)...)
			}
		}
		return scopes
	default:
		return nil
	}
}

func scopesIntersect(granted []string, allowed []string) bool {
	for _, grantedScope := range granted {
		for _, allowedScope := range allowed {
			if grantedScope == allowedScope {
				return true
			}
		}
	}
	return false
}

func scopeDeniedFailure() *VerifyRequestResult {
	return &VerifyRequestResult{
		IsOk: false,
		ClientError: ft.NewAuthorizationError(
			ft.ErrorKey("err_insufficient_token_scope"),
			"The access token does not grant any scope this endpoint accepts.",
		),
	}
}
