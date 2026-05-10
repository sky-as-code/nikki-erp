package modelmetadata

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
)

type ModelMetadataRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys models.ModelMetadata) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []models.ModelMetadata) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, src models.ModelMetadata) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[models.ModelMetadata], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[models.ModelMetadata]], error)
	Update(ctx corectx.Context, src models.ModelMetadata) (*dyn.OpResult[dyn.MutateResultData], error)
}
