package app

import (
	it "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/login"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

func NewLoginApplicationServiceImpl(loginSvc it.LoginDomainService) it.LoginAppService {
	return &LoginApplicationServiceImpl{loginSvc: loginSvc}
}

type LoginApplicationServiceImpl struct {
	loginSvc it.LoginDomainService
}

func (this *LoginApplicationServiceImpl) Authenticate(ctx corectx.Context, cmd it.AuthenticateCommand) (result *it.AuthenticateResult, err error) {
	return this.loginSvc.Authenticate(ctx, cmd)
}

func (this *LoginApplicationServiceImpl) RefreshToken(ctx corectx.Context, cmd it.RefreshTokenCommand) (result *it.RefreshTokenResult, err error) {
	return this.loginSvc.RefreshToken(ctx, cmd)
}
