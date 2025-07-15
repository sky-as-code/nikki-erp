package login

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
)

type AttemptService interface {
	CreateLoginAttempt(ctx context.Context, cmd CreateLoginAttemptCommand) (result *CreateLoginAttemptResult, err error)
	GetAttemptById(ctx context.Context, query GetAttemptByIdQuery) (result *GetAttemptByIdResult, err error)
	UpdateLoginAttempt(ctx context.Context, cmd UpdateLoginAttemptCommand) (result *UpdateLoginAttemptResult, err error)
}

type LoginService interface {
	Authenticate(ctx context.Context, cmd AuthenticateCommand) (result *AuthenticateResult, err error)
}

type AttemptRepository interface {
	Create(ctx context.Context, attempt domain.LoginAttempt) (*domain.LoginAttempt, error)
	Update(ctx context.Context, attempt domain.LoginAttempt) (*domain.LoginAttempt, error)
	FindById(ctx context.Context, param FindByIdParam) (*domain.LoginAttempt, error)
}

type FindByIdParam = GetAttemptByIdQuery

type LoginParam struct {
	SubjectType domain.SubjectType `json:"subjectType"`
	Username    string             `json:"username"`
	Password    string             `json:"password"`
}

type LoginMethod interface {
	Execute(ctx context.Context, param LoginParam) (bool, error)
	Name() string
}
