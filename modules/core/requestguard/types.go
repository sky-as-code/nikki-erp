package requestguard

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type RequestGuardService interface {
	// Calculate a fingerprint for the request that can be used to identify the request.
	// This can be used for caching or to prevent replay attacks.
	CalcRequestFingerprint(ctx corectx.Context, request *http.Request) (fingerprint string, err error)
	GetCorsMiddleware(ctx corectx.Context) (echo.MiddlewareFunc, error)
	GetUserEntitlements(ctx corectx.Context, query GetUserEntitlementsQuery) (*GetUserEntitlementsResult, error)
	VerifyJwt(ctx corectx.Context, request *http.Request) (result *VerifyRequestResult, err error)
}

type VerifyRequestResult struct {
	IsOk        bool
	JwtClaims   jwt.Claims
	ClientError *ft.ClientErrorItem
}

type GetUserEntitlementsQuery = ExtGetUserEntitlementsQuery
type GetUserEntitlementsResultData = ExtGetUserEntitlementsResultData
type GetUserEntitlementsResult = dyn.OpResult[GetUserEntitlementsResultData]

type ResourceScope string

const (
	ResourceScopeDomain  = ResourceScope("domain")
	ResourceScopeOrg     = ResourceScope("org")
	ResourceScopeOrgUnit = ResourceScope("orgunit")
	ResourceScopePrivate = ResourceScope("private")
)
