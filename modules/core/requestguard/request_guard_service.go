package requestguard

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/core/config"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
)

type StaticRequestGuardServiceParams struct {
	dig.In

	ConfigSvc config.ConfigService
}

func NewStaticRequestGuardServiceImpl(params StaticRequestGuardServiceParams) RequestGuardService {
	return &StaticRequestGuardServiceImpl{
		configSvc: params.ConfigSvc,
	}
}

type StaticRequestGuardServiceImpl struct {
	configSvc      config.ConfigService
	corsMiddleware echo.MiddlewareFunc
}

func (this *StaticRequestGuardServiceImpl) VerifyRequest(request *http.Request) (result *VerifyRequestResult, err error) {
	cfg := this.configSvc

	if cfg.GetBool(c.RequestGuardMtlsEnabled) {
		if result, err = this.verifySecuredConnection(request); result != nil || err != nil {
			return
		}
	}
	if cfg.GetBool(c.RequestGuardAccessTokenEnabled) {
		if result, err = this.verifyJwt(request); result != nil || err != nil {
			return
		}
	}
	if cfg.GetBool(c.RequestGuardSessionBlacklistEnabled) {
		if result, err = this.verifySessionBlacklist(request); result != nil || err != nil {
			return
		}
	}
	return &VerifyRequestResult{
		IsOk:        true,
		HttpStatus:  http.StatusOK,
		ClientError: nil,
	}, nil
}

func (this *StaticRequestGuardServiceImpl) GetCorsMiddleware() (echo.MiddlewareFunc, error) {
	if !this.configSvc.GetBool(c.RequestGuardCorsEnabled) {
		return nil, nil
	}

	if this.corsMiddleware == nil {
		this.corsMiddleware = middleware.CORSWithConfig(this.configCors())
	}
	return this.corsMiddleware, nil
}

func (this *StaticRequestGuardServiceImpl) configCors() middleware.CORSConfig {
	corsOrigins := this.configSvc.GetStrArr(c.HttpCorsOrigins)
	corsHeaders := this.configSvc.GetStrArr(c.HttpCorsHeaders, "")
	if len(corsHeaders) == 0 {
		corsHeaders = []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization}
	}
	corsMethods := this.configSvc.GetStrArr(c.HttpCorsMethods, "")
	if len(corsMethods) == 0 {
		corsMethods = []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE}
	}

	return middleware.CORSConfig{
		// TODO: Allow config CORS from database
		AllowOrigins: corsOrigins,
		AllowHeaders: corsHeaders,
		AllowMethods: corsMethods,
	}
}

func (this *StaticRequestGuardServiceImpl) verifySecuredConnection(request *http.Request) (*VerifyRequestResult, error) {
	return nil, nil
}

func (this *StaticRequestGuardServiceImpl) verifyJwt(request *http.Request) (*VerifyRequestResult, error) {
	if this.configSvc.GetBool(c.RequestGuardAccessTokenDpopEnabled) {
		result, err := this.verifyJwtDpop(request)
		if result != nil || err != nil {
			return result, err
		}
	}
	return nil, nil
}

// Verify JWT DPoP (OAuth2 Demonstraing Proof of Possession)
func (this *StaticRequestGuardServiceImpl) verifyJwtDpop(request *http.Request) (*VerifyRequestResult, error) {

	return nil, nil
}

func (this *StaticRequestGuardServiceImpl) verifySessionBlacklist(request *http.Request) (*VerifyRequestResult, error) {

	return nil, nil
}
