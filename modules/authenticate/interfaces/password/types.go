package password

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
)

type PasswordService interface {
	CreateOtpPassword(ctx context.Context, cmd CreateOtpPasswordCommand) (*CreateOtpPasswordResult, error)
	ConfirmOtpPassword(ctx context.Context, cmd ConfirmOtpPasswordCommand) (*ConfirmOtpPasswordResult, error)
	CreateTempPassword(ctx context.Context, cmd CreateTempPasswordCommand) (*CreateTempPasswordResult, error)
	SetPassword(ctx context.Context, cmd SetPasswordCommand) (*SetPasswordResult, error)
	VerifyPassword(ctx context.Context, cmd VerifyPasswordQuery) (*VerifyPasswordResult, error)
	VerifyOtpCode(ctx context.Context, cmd VerifyOtpCodeQuery) (*VerifyOtpCodeResult, error)
}

type PasswordStoreRepository interface {
	Create(ctx context.Context, store domain.PasswordStore) (*domain.PasswordStore, error)
	Update(ctx context.Context, store domain.PasswordStore) (*domain.PasswordStore, error)
	FindBySubject(ctx context.Context, param FindBySubjectParam) (*domain.PasswordStore, error)
}

type FindBySubjectParam struct {
	SubjectType domain.SubjectType `json:"subjectType"`
	SubjectRef  model.Id           `json:"subjectRef"`
}
