package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/util"
	it "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/login"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
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
	var request StartLoginFlowRequest
	if err := echoCtx.Bind(&request); err != nil {
		return err
	}

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
	reqCtx := echoCtx.Request().Context().(crud.Context)
	result, err := this.attemptSvc.CreateLoginAttempt(reqCtx, cmd)

	if err != nil {
		return err
	}

	if result.ClientError != nil {
		return httpserver.JsonBadRequest(echoCtx, *result.ClientError)
	}

	response := NewStartLoginFlowResponse(*result)
	return httpserver.JsonCreated(echoCtx, response)
}

func (this LoginRest) Authenticate(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST authenticate"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.loginSvc.Authenticate,
		func(request AuthenticateRequest) it.AuthenticateCommand {
			return it.AuthenticateCommand(request)
		},
		func(result it.AuthenticateResult) AuthenticateResponse {
			return AuthenticateResponse(*result.Data)
		},
		httpserver.JsonOk,
	)
	return err
}
