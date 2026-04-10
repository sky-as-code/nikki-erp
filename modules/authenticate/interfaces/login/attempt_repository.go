package login

import (
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type AttemptRepository interface {
	dyn.DynamicModelRepository
	Insert(ctx corectx.Context, attempt domain.LoginAttempt) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.LoginAttempt], error)
	Update(ctx corectx.Context, attempt domain.LoginAttempt) (*dyn.OpResult[dyn.MutateResultData], error)
}
