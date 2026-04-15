package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/util"
	it "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/login"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
)

type loginRestParams struct {
	dig.In

	AttemptSvc it.AttemptService
	LoginSvc   it.LoginService
}

func NewLoginRest(params loginRestParams) *LoginRest {
	return &LoginRest{
		attemptSvc: params.AttemptSvc,
		loginSvc:   params.LoginSvc,
	}
}

type LoginRest struct {
	httpserver.RestBase
	attemptSvc it.AttemptService
	loginSvc   it.LoginService
}

func (this LoginRest) StartLoginFlow(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST start login flow"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.attemptSvc.CreateLoginAttempt,
		func(request StartLoginFlowRequest) it.CreateLoginAttemptCommand {
			deviceName := echoCtx.Request().Header.Get("User-Agent")
			if deviceName == "" && request.DeviceName != nil && len(*request.DeviceName) > 0 {
				deviceName = *request.DeviceName
			}

			cmd := it.NewCreateLoginAttemptCommand()
			cmd.SetDeviceIp(util.ToPtr(echoCtx.RealIP()))
			cmd.SetDeviceName(&deviceName)
			cmd.SetDeviceLocation(util.ToPtr(echoCtx.RealIP())) // TODO: Use geoip service to get location
			cmd.SetPrincipalType(request.PrincipalType)
			cmd.SetUsername(&request.Username)
			return cmd
		},
		NewStartLoginFlowResponse,
		httpserver.JsonCreated,
	)
}

func (this LoginRest) Authenticate(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST authenticate"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.loginSvc.Authenticate,
		func(request AuthenticateRequest) it.AuthenticateCommand {
			return it.AuthenticateCommand(request)
		},
		func(data it.AuthenticateResultData) AuthenticateResponse {
			return AuthenticateResponse(data)
		},
		httpserver.JsonOk,
	)
}

func (this LoginRest) RefreshToken(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST refresh token"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.loginSvc.RefreshToken,
		func(request RefreshTokenRequest) it.RefreshTokenCommand {
			return it.RefreshTokenCommand(request)
		},
		func(data it.RefreshTokenResultData) RefreshTokenResponse {
			return RefreshTokenResponse(data)
		},
		httpserver.JsonOk,
	)
}
