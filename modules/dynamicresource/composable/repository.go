package composable

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
)

// CrudRepository reads and writes the resource records through the SQL query builder.
//
// A module extends it by embedding the default implementation into its own
// {Resource}RepositoryImpl and handing that struct to DynamicResourceEngineOnionImpl.NewRepositoryFn.
type CrudRepository interface {
	// Embedding this interface lets the repository be passed directly to the generic helpers
	// of modules/core/dynamicmodel/crud, which expect a dyn.DynamicModelRepository.
	dyn.DynamicModelRepository

	BeginTransaction(ctx corectx.Context) (database.DbTransaction, error)

	Insert(ctx corectx.Context, data dmodel.DynamicFields) (*dyn.OpResult[int], error)
	Update(ctx corectx.Context, data dmodel.DynamicFields) (*MutateResult, error)
	DeleteOne(ctx corectx.Context, keys dmodel.DynamicFields) (*MutateResult, error)

	// FindByKeys fetches the single record identified by the given primary or unique keys.
	FindByKeys(ctx corectx.Context, keys dmodel.DynamicFields) (*dyn.OpResult[dmodel.DynamicFields], error)

	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[dmodel.DynamicFields], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*SearchResult, error)
	Exists(ctx corectx.Context, keys []dmodel.DynamicFields) (*dyn.OpResult[dyn.RepoExistsResult], error)

	Schema() *dmodel.ModelSchema
}

// NewRepositoryParam carries the core services a repository needs. The onion fills it from
// BuildParam for the schema it serves.
type NewRepositoryParam struct {
	Client        orm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  orm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
	Schema        *dmodel.ModelSchema
}

func NewDefaultCrudRepository(param NewRepositoryParam) CrudRepository {
	dynamicRepo := param.NewBaseRepoFn(
		dyn.NewBaseRepoParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       param.Schema,
		},
	)
	return &DefaultCrudRepositoryImpl{dynamicRepo: dynamicRepo}
}

// DefaultCrudRepositoryImpl is the schema-agnostic counterpart of the hand-written
// repositories such as iam's UserDynamicRepository: it delegates to the same baserepo helpers,
// but speaks DynamicFields instead of a typed domain model.
type DefaultCrudRepositoryImpl struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *DefaultCrudRepositoryImpl) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *DefaultCrudRepositoryImpl) Schema() *dmodel.ModelSchema {
	return this.dynamicRepo.Schema()
}

func (this *DefaultCrudRepositoryImpl) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}

func (this *DefaultCrudRepositoryImpl) Insert(
	ctx corectx.Context, data dmodel.DynamicFields,
) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, NewDynamicEntityFrom(data))
}

func (this *DefaultCrudRepositoryImpl) Update(
	ctx corectx.Context, data dmodel.DynamicFields,
) (*MutateResult, error) {
	return baserepo.Update(ctx, this.dynamicRepo, data)
}

func (this *DefaultCrudRepositoryImpl) DeleteOne(
	ctx corectx.Context, keys dmodel.DynamicFields,
) (*MutateResult, error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys)
}

// FindByKeys reads every column of the schema.
func (this *DefaultCrudRepositoryImpl) FindByKeys(
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

func (this *DefaultCrudRepositoryImpl) GetOne(
	ctx corectx.Context, param dyn.RepoGetOneParam,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	return this.dynamicRepo.GetOne(ctx, param)
}

func (this *DefaultCrudRepositoryImpl) Search(
	ctx corectx.Context, param dyn.RepoSearchParam,
) (*SearchResult, error) {
	return this.dynamicRepo.Search(ctx, param)
}

func (this *DefaultCrudRepositoryImpl) Exists(
	ctx corectx.Context, keys []dmodel.DynamicFields,
) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	return baserepo.Exists(ctx, this.dynamicRepo, keys)
}
