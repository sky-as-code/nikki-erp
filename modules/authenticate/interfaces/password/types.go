package password

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
)

type PasswordService interface {
	SetPassword(ctx context.Context, cmd SetPasswordCommand) (result *SetPasswordResult, err error)
	IsPasswordMatched(ctx context.Context, cmd IsPasswordMatchedQuery) (result *IsPasswordMatchedResult, err error)
}

type PasswordStoreRepository interface {
	Create(ctx context.Context, attempt domain.PasswordStore) (*domain.PasswordStore, error)
	Update(ctx context.Context, attempt domain.PasswordStore) (*domain.PasswordStore, error)
	FindBySubject(ctx context.Context, param FindBySubjectParam) (*domain.PasswordStore, error)
}

type FindBySubjectParam struct {
	SubjectType domain.SubjectType `json:"subjectType"`
	SubjectRef  model.Id           `json:"subjectRef"`
}
