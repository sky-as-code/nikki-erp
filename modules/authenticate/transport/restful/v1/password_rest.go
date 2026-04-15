package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	it "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/password"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
)

type passwordRestParam struct {
	dig.In

	PasswordSvc it.PasswordService
}

func NewPasswordRest(params passwordRestParam) *PasswordRest {
	return &PasswordRest{
		passwordSvc: params.PasswordSvc,
	}
}

type PasswordRest struct {
	httpserver.RestBase
	passwordSvc it.PasswordService
}

func (this PasswordRest) CreatePasswordOtp(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create password OTP"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.passwordSvc.CreatePasswordOtp,
		func(request CreateOtpPasswordRequest) it.CreatePasswordOtpCommand {
			return it.CreatePasswordOtpCommand(request)
		},
		NewCreateOtpPasswordResponse,
		httpserver.JsonOk,
	)
}

func (this PasswordRest) ConfirmPasswordOtp(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST confirm password OTP"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.passwordSvc.ConfirmPasswordOtp,
		func(request ConfirmOtpPasswordRequest) it.ConfirmPasswordOtpCommand {
			return it.ConfirmPasswordOtpCommand(request)
		},
		NewConfirmOtpPasswordResponse,
		httpserver.JsonOk,
	)
}

func (this PasswordRest) CreatePasswordTemp(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create password temp"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.passwordSvc.CreatePasswordTemp,
		func(request CreateTempPasswordRequest) it.CreatePasswordTempCommand {
			return it.CreatePasswordTempCommand(request)
		},
		NewCreateTempPasswordResponse,
		httpserver.JsonOk,
	)
}

func (this PasswordRest) SetPassword(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST set password"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.passwordSvc.SetPassword,
		func(request SetPasswordRequest) it.SetPasswordCommand {
			return it.SetPasswordCommand(request)
		},
		NewSetPasswordResponse,
		httpserver.JsonOk,
	)
}
