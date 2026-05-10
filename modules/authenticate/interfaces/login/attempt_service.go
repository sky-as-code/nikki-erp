package login

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type AttemptDomainService interface {
	CreateLoginAttempt(ctx corectx.Context, cmd CreateLoginAttemptCommand) (result *CreateLoginAttemptResult, err error)
	GetAttempt(ctx corectx.Context, query GetAttemptQuery) (result *GetAttemptResult, err error)
	UpdateLoginAttempt(ctx corectx.Context, cmd UpdateLoginAttemptCommand) (result *UpdateLoginAttemptResult, err error)
}

type AttemptAppService interface {
	CreateLoginAttempt(ctx corectx.Context, cmd CreateLoginAttemptCommand) (result *CreateLoginAttemptResult, err error)
	GetAttempt(ctx corectx.Context, query GetAttemptQuery) (result *GetAttemptResult, err error)
	UpdateLoginAttempt(ctx corectx.Context, cmd UpdateLoginAttemptCommand) (result *UpdateLoginAttemptResult, err error)
}
