package engine

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/baserepo"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	it "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// NewRepositoryParam carries the core services a repository needs.
// The registry fills it once and reuses it for every engine it builds.
type NewRepositoryParam struct {
	Client        orm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  orm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
	Schema        *dmodel.ModelSchema
}

func NewDynamicResourceRepository(param NewRepositoryParam) it.DynamicResourceRepository {
	dynamicRepo := param.NewBaseRepoFn(
		dyn.NewBaseRepoParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       param.Schema,
		},
	)
	return &DynamicResourceRepositoryImpl{dynamicRepo: dynamicRepo}
}

// DynamicResourceRepositoryImpl is the schema-agnostic counterpart of the hand-written
// repositories such as iam's UserDynamicRepository: it delegates to the same baserepo
// helpers, but speaks DynamicFields instead of a typed domain model.
type DynamicResourceRepositoryImpl struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *DynamicResourceRepositoryImpl) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *DynamicResourceRepositoryImpl) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}

func (this *DynamicResourceRepositoryImpl) Insert(
	ctx corectx.Context, data dmodel.DynamicFields,
) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, it.NewDynamicEntityFrom(data))
}

func (this *DynamicResourceRepositoryImpl) Update(
	ctx corectx.Context, data dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, data)
}

func (this *DynamicResourceRepositoryImpl) DeleteOne(
	ctx corectx.Context, keys dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys)
}

// FindByKeys fetches the single record identified by the given keys, reading every
// column of the schema.
func (this *DynamicResourceRepositoryImpl) FindByKeys(
	ctx corectx.Context, keys dmodel.DynamicFields,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	return this.GetOne(ctx, dyn.RepoGetOneParam{
		Filter: keys,
		Fields: columnNames(this.dynamicRepo.Schema()),
	})
}

// columnNames lists the schema fields a client may select and receive, which is what the default
// projection must be. Virtual scalars are included: they have no database column but are filled
// by a service after the read, so omitting them here would make them unreachable by default.
func columnNames(schema *dmodel.ModelSchema) []string {
	return array.Map(schema.ReadableFields(), func(field *dmodel.ModelField) string {
		return field.Name()
	})
}

func (this *DynamicResourceRepositoryImpl) GetOne(
	ctx corectx.Context, param dyn.RepoGetOneParam,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	return this.dynamicRepo.GetOne(ctx, param)
}

func (this *DynamicResourceRepositoryImpl) Search(
	ctx corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error) {
	return this.dynamicRepo.Search(ctx, param)
}

func (this *DynamicResourceRepositoryImpl) Exists(
	ctx corectx.Context, keys []dmodel.DynamicFields,
) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	return baserepo.Exists(ctx, this.dynamicRepo, keys)
}
