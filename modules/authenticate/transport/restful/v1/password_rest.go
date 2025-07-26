package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	it "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/password"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
)

type passwordRestParams struct {
	dig.In

	PasswordSvc it.PasswordService
}

func NewPasswordRest(params passwordRestParams) *PasswordRest {
	return &PasswordRest{
		passwordSvc: params.PasswordSvc,
	}
}

type PasswordRest struct {
	httpserver.RestBase
	passwordSvc it.PasswordService
}

func (this PasswordRest) CreateOtpPassword(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create otp password"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.passwordSvc.CreateOtpPassword,
		func(request CreateOtpPasswordRequest) it.CreateOtpPasswordCommand {
			return it.CreateOtpPasswordCommand(request)
		},
		NewCreateOtpPasswordResponse,
		httpserver.JsonOk,
	)
	return err
}

func (this PasswordRest) ConfirmOtpPassword(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST confirm otp password"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.passwordSvc.ConfirmOtpPassword,
		func(request ConfirmOtpPasswordRequest) it.ConfirmOtpPasswordCommand {
			return it.ConfirmOtpPasswordCommand(request)
		},
		NewConfirmOtpPasswordResponse,
		httpserver.JsonOk,
	)
	return err
}

func (this PasswordRest) CreateTempPassword(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create temp password"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.passwordSvc.CreateTempPassword,
		func(request CreateTempPasswordRequest) it.CreateTempPasswordCommand {
			return it.CreateTempPasswordCommand(request)
		},
		NewCreateTempPasswordResponse,
		httpserver.JsonOk,
	)
	return err
}

func (this PasswordRest) SetPassword(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST set password"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.passwordSvc.SetPassword,
		func(request SetPasswordRequest) it.SetPasswordCommand {
			return it.SetPasswordCommand(request)
		},
		NewSetPasswordResponse,
		httpserver.JsonOk,
	)
	return err
}
