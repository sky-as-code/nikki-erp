package login

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type AttemptService interface {
	CreateLoginAttempt(ctx corectx.Context, cmd CreateLoginAttemptCommand) (result *CreateLoginAttemptResult, err error)
	GetAttemptById(ctx corectx.Context, query GetAttemptByIdQuery) (result *GetAttemptByIdResult, err error)
	UpdateLoginAttempt(ctx corectx.Context, cmd UpdateLoginAttemptCommand) (result *UpdateLoginAttemptResult, err error)
}

type LoginService interface {
	Authenticate(ctx corectx.Context, cmd AuthenticateCommand) (result *AuthenticateResult, err error)
	RefreshToken(ctx corectx.Context, cmd RefreshTokenCommand) (result *RefreshTokenResult, err error)
}

type AttemptRepository interface {
	dyn.DynamicModelRepository
	Insert(ctx corectx.Context, attempt domain.LoginAttempt) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.LoginAttempt], error)
	Update(ctx corectx.Context, attempt domain.LoginAttempt) (*dyn.OpResult[dyn.MutateResultData], error)
}

type FindByIdParam = GetAttemptByIdQuery

type LoginParam struct {
	SubjectType domain.SubjectType `json:"subject_type"`
	Username    string             `json:"username"`
	Password    string             `json:"password"`
}

type LoginMethod interface {
	Execute(ctx corectx.Context, param LoginParam) (*ExecuteResult, error)
	Name() string
	SkipMethod() *SkippedMethod
}

type ExecuteResult struct {
	IsVerified   bool
	FailedReason string
	ClientErrors ft.ClientErrors
}

type SkippedMethod string

const (
	SkippedMethodAll      SkippedMethod = "*"
	SkippedMethodPassword SkippedMethod = "password"
)
