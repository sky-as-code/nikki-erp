package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/login"
)

func NewAttemptApplicationServiceImpl(attemptSvc it.AttemptDomainService) it.AttemptAppService {
	return &AttemptApplicationServiceImpl{attemptSvc: attemptSvc}
}

type AttemptApplicationServiceImpl struct {
	attemptSvc it.AttemptDomainService
}

func (this *AttemptApplicationServiceImpl) CreateLoginAttempt(ctx corectx.Context, cmd it.CreateLoginAttemptCommand) (result *it.CreateLoginAttemptResult, err error) {
	return this.attemptSvc.CreateLoginAttempt(ctx, cmd)
}

func (this *AttemptApplicationServiceImpl) GetAttempt(ctx corectx.Context, query it.GetAttemptQuery) (result *it.GetAttemptResult, err error) {
	return this.attemptSvc.GetAttempt(ctx, query)
}

func (this *AttemptApplicationServiceImpl) UpdateLoginAttempt(ctx corectx.Context, cmd it.UpdateLoginAttemptCommand) (result *it.UpdateLoginAttemptResult, err error) {
	return this.attemptSvc.UpdateLoginAttempt(ctx, cmd)
}
