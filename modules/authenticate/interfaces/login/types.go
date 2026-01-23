package login

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type AttemptService interface {
	CreateLoginAttempt(ctx crud.Context, cmd CreateLoginAttemptCommand) (result *CreateLoginAttemptResult, err error)
	GetAttemptById(ctx crud.Context, query GetAttemptByIdQuery) (result *GetAttemptByIdResult, err error)
	UpdateLoginAttempt(ctx crud.Context, cmd UpdateLoginAttemptCommand) (result *UpdateLoginAttemptResult, err error)
}

type LoginService interface {
	Authenticate(ctx crud.Context, cmd AuthenticateCommand) (result *AuthenticateResult, err error)
	RefreshToken(ctx crud.Context, cmd RefreshTokenCommand) (result *RefreshTokenResult, err error)
}

type AttemptRepository interface {
	Create(ctx crud.Context, attempt domain.LoginAttempt) (*domain.LoginAttempt, error)
	Update(ctx crud.Context, attempt domain.LoginAttempt) (*domain.LoginAttempt, error)
	FindById(ctx crud.Context, param FindByIdParam) (*domain.LoginAttempt, error)
}

type FindByIdParam = GetAttemptByIdQuery

type LoginParam struct {
	SubjectType domain.SubjectType `json:"subjectType"`
	Username    string             `json:"username"`
	Password    string             `json:"password"`
}

type LoginMethod interface {
	Execute(ctx crud.Context, param LoginParam) (*ExecuteResult, error)
	Name() string
	SkipMethod() *SkippedMethod
}

type ExecuteResult struct {
	IsVerified   bool
	FailedReason string
	ClientErr    *ft.ClientError
}

type SkippedMethod string

const (
	SkippedMethodAll      SkippedMethod = "*"
	SkippedMethodPassword SkippedMethod = "password"
)
