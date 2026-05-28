package resource

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
)

type ResourceRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys models.Resource) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []models.Resource) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, row models.Resource) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[models.Resource], error)
	GetByAction(ctx corectx.Context, param RepoGetByActionParam) (*dyn.OpResult[models.Resource], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[models.Resource]], error)
	Update(ctx corectx.Context, row models.Resource) (*dyn.OpResult[dyn.MutateResultData], error)
}

type RepoGetByActionParam struct {
	ActionCode string `json:"action_code"`
	Columns    []string
}
