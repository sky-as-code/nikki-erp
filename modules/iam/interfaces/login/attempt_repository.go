package login

import (
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type AttemptRepository interface {
	dyn.DynamicModelRepository
	Insert(ctx corectx.Context, attempt models.LoginAttempt) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[models.LoginAttempt], error)
	Update(ctx corectx.Context, attempt models.LoginAttempt) (*dyn.OpResult[dyn.MutateResultData], error)
}
