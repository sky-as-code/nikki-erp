package enum

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
)

type EnumRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys models.Enum) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []models.Enum) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, enum models.Enum) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[models.Enum], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[models.Enum]], error)
	Update(ctx corectx.Context, enum models.Enum) (*dyn.OpResult[dyn.MutateResultData], error)
}
