package password

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type PasswordService interface {
	CreateOtpPassword(ctx corectx.Context, cmd CreateOtpPasswordCommand) (*CreateOtpPasswordResult, error)
	ConfirmOtpPassword(ctx corectx.Context, cmd ConfirmOtpPasswordCommand) (*ConfirmOtpPasswordResult, error)
	CreateTempPassword(ctx corectx.Context, cmd CreateTempPasswordCommand) (*CreateTempPasswordResult, error)
	SetPassword(ctx corectx.Context, cmd SetPasswordCommand) (*SetPasswordResult, error)
	VerifyPassword(ctx corectx.Context, cmd VerifyPasswordQuery) (*VerifyPasswordResult, error)
	VerifyOtpCode(ctx corectx.Context, cmd VerifyOtpCodeQuery) (*VerifyOtpCodeResult, error)
}

type PasswordStoreRepository interface {
	Insert(ctx corectx.Context, store domain.PasswordStore) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.PasswordStore], error)
	Update(ctx corectx.Context, store domain.PasswordStore) (*dyn.OpResult[dyn.MutateResultData], error)
}

type FindBySubjectParam struct {
	SubjectType domain.SubjectType `json:"subject_type"`
	SubjectRef  model.Id           `json:"subject_ref"`
}
