package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/password"
)

func NewPasswordApplicationServiceImpl(passwordSvc it.PasswordDomainService) it.PasswordAppService {
	return &PasswordApplicationServiceImpl{passwordSvc: passwordSvc}
}

type PasswordApplicationServiceImpl struct {
	passwordSvc it.PasswordDomainService
}

func (this *PasswordApplicationServiceImpl) CreatePasswordOtp(ctx corectx.Context, cmd it.CreatePasswordOtpCommand) (*it.CreatePasswordOtpResult, error) {
	return this.passwordSvc.CreatePasswordOtp(ctx, cmd)
}

func (this *PasswordApplicationServiceImpl) ConfirmPasswordOtp(ctx corectx.Context, cmd it.ConfirmPasswordOtpCommand) (*it.ConfirmPasswordOtpResult, error) {
	return this.passwordSvc.ConfirmPasswordOtp(ctx, cmd)
}

func (this *PasswordApplicationServiceImpl) CreatePasswordTemp(ctx corectx.Context, cmd it.CreatePasswordTempCommand) (*it.CreatePasswordTempResult, error) {
	return this.passwordSvc.CreatePasswordTemp(ctx, cmd)
}

func (this *PasswordApplicationServiceImpl) SetPassword(ctx corectx.Context, cmd it.SetPasswordCommand) (*it.SetPasswordResult, error) {
	return this.passwordSvc.SetPassword(ctx, cmd)
}

func (this *PasswordApplicationServiceImpl) VerifyPassword(ctx corectx.Context, cmd it.VerifyPasswordQuery) (*it.VerifyPasswordResult, error) {
	return this.passwordSvc.VerifyPassword(ctx, cmd)
}

func (this *PasswordApplicationServiceImpl) VerifyOtpCode(ctx corectx.Context, cmd it.VerifyPasswordOtpQuery) (*it.VerifyOtpCodeResult, error) {
	return this.passwordSvc.VerifyOtpCode(ctx, cmd)
}
