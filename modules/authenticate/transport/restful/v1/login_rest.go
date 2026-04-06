package v1

import (
	"github.com/labstack/echo/v4"
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

func (this LoginRest) StartLoginFlow(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST start login flow"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.attemptSvc.CreateLoginAttempt,
		func(request StartLoginFlowRequest) it.CreateLoginAttemptCommand {
			var deviceName *string
			deviceName = util.ToPtr(echoCtx.Request().Header.Get("User-Agent"))
			if len(*deviceName) == 0 && request.DeviceName != nil && len(*request.DeviceName) > 0 {
				deviceName = request.DeviceName
			}

			cmd := it.CreateLoginAttemptCommand{
				DeviceIp:       util.ToPtr(echoCtx.RealIP()),
				DeviceName:     deviceName,
				DeviceLocation: util.ToPtr(echoCtx.RealIP()), // Use geoip service to get location
				SubjectType:    request.SubjectType,
				Username:       request.Username,
			}
			return cmd
		},
		func(data *it.CreateLoginAttemptResultData) StartLoginFlowResponse {
			return NewStartLoginFlowResponse(data)
		},
		httpserver.JsonCreated,
	)
}

func (this LoginRest) Authenticate(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST authenticate"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest2(
		echoCtx,
		this.loginSvc.Authenticate,
		func(request AuthenticateRequest) it.AuthenticateCommand {
			return it.AuthenticateCommand(request)
		},
		func(data *it.AuthenticateResultData) AuthenticateResponse {
			if data == nil {
				return AuthenticateResponse{}
			}
			return AuthenticateResponse(*data)
		},
		httpserver.JsonOk,
	)
	return err
}

func (this LoginRest) RefreshToken(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST refresh token"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest2(
		echoCtx,
		this.loginSvc.RefreshToken,
		func(request RefreshTokenRequest) it.RefreshTokenCommand {
			return it.RefreshTokenCommand(request)
		},
		func(data *it.RefreshTokenResultData) RefreshTokenResponse {
			if data == nil {
				return RefreshTokenResponse{}
			}
			return RefreshTokenResponse(*data)
		},
		httpserver.JsonOk,
	)
	return err
}
