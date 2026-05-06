package tag

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
)

type TagRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys models.Tag) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []models.Tag) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, tag models.Tag) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[models.Tag], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[models.Tag]], error)
	Update(ctx corectx.Context, tag models.Tag) (*dyn.OpResult[dyn.MutateResultData], error)
}
