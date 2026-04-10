package requestguard

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

type RequestGuardService interface {
	// Calculate a fingerprint for the request that can be used to identify the request.
	// This can be used for caching or to prevent replay attacks.
	CalcRequestFingerprint(request *http.Request) (fingerprint string, err error)
	VerifyTrustedConnection(request *http.Request) (result *VerifyRequestResult, err error)
	VerifyJwt(request *http.Request) (result *VerifyRequestResult, err error)
	GetCorsMiddleware() (echo.MiddlewareFunc, error)
}

type VerifyRequestResult struct {
	IsOk        bool
	JwtClaims   jwt.Claims
	HttpStatus  int
	ClientError *ft.ClientErrorItem
}
