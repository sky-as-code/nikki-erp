package module

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain"
)

type ModuleRepository interface {
	dyn.DynamicModelRepository
	AcquireLock(ctx corectx.Context) (bool, error)
	ReleaseLock(ctx corectx.Context) error
	DeleteOne(ctx corectx.Context, keys domain.ModuleMetadata) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []domain.ModuleMetadata) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, module domain.ModuleMetadata) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.ModuleMetadata], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.ModuleMetadata]], error)
	Update(ctx corectx.Context, module domain.ModuleMetadata) (*dyn.OpResult[dyn.MutateResultData], error)
}
