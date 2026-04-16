package fieldmetadata

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain"
)

type FieldMetadataRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys domain.FieldMetadata) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []domain.FieldMetadata) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, src domain.FieldMetadata) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.FieldMetadata], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.FieldMetadata]], error)
	Update(ctx corectx.Context, src domain.FieldMetadata) (*dyn.OpResult[dyn.MutateResultData], error)
}
