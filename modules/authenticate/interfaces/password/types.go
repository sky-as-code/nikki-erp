package password

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type PasswordService interface {
	CreateOtpPassword(ctx crud.Context, cmd CreateOtpPasswordCommand) (*CreateOtpPasswordResult, error)
	ConfirmOtpPassword(ctx crud.Context, cmd ConfirmOtpPasswordCommand) (*ConfirmOtpPasswordResult, error)
	CreateTempPassword(ctx crud.Context, cmd CreateTempPasswordCommand) (*CreateTempPasswordResult, error)
	SetPassword(ctx crud.Context, cmd SetPasswordCommand) (*SetPasswordResult, error)
	VerifyPassword(ctx crud.Context, cmd VerifyPasswordQuery) (*VerifyPasswordResult, error)
	VerifyOtpCode(ctx crud.Context, cmd VerifyOtpCodeQuery) (*VerifyOtpCodeResult, error)
}

type PasswordStoreRepository interface {
	Create(ctx crud.Context, store domain.PasswordStore) (*domain.PasswordStore, error)
	Update(ctx crud.Context, store domain.PasswordStore) (*domain.PasswordStore, error)
	FindBySubject(ctx crud.Context, param FindBySubjectParam) (*domain.PasswordStore, error)
}

type FindBySubjectParam struct {
	SubjectType domain.SubjectType `json:"subjectType"`
	SubjectRef  model.Id           `json:"subjectRef"`
}
