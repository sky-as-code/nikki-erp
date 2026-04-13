package requestguard

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type RequestGuardService interface {
	// Calculate a fingerprint for the request that can be used to identify the request.
	// This can be used for caching or to prevent replay attacks.
	CalcRequestFingerprint(ctx corectx.Context, request *http.Request) (fingerprint string, err error)
	VerifyTrustedConnection(ctx corectx.Context, request *http.Request) (result *VerifyRequestResult, err error)
	VerifyJwt(ctx corectx.Context, request *http.Request) (result *VerifyRequestResult, err error)
	GetCorsMiddleware(ctx corectx.Context) (echo.MiddlewareFunc, error)
}

type VerifyRequestResult struct {
	IsOk      bool
	JwtClaims jwt.Claims
	// HttpStatus  int
	ClientError *ft.ClientErrorItem
}
