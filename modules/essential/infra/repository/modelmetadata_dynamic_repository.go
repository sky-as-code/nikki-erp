package repository

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/array"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/baserepo"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/modelmetadata"
)

type ModelMetadataDynamicRepositoryParam struct {
	dig.In
	Client        orm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  orm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewModelMetadataDynamicRepository(param ModelMetadataDynamicRepositoryParam) it.ModelMetadataRepository {
	dynamicRepo := param.NewBaseRepoFn(dyn.NewBaseRepoParam{
		Client: param.Client, ConfigSvc: param.ConfigSvc, QueryBuilder: param.QueryBuilder,
		Logger: param.Logger, Schema: dmodel.MustGetSchema(domain.ModelMetadataSchemaName),
	})
	return &ModelMetadataDynamicRepository{dynamicRepo: dynamicRepo}
}

type ModelMetadataDynamicRepository struct{ dynamicRepo dyn.BaseDynamicRepository }

func (this *ModelMetadataDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}
func (this *ModelMetadataDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}
func (this *ModelMetadataDynamicRepository) DeleteOne(ctx corectx.Context, keys domain.ModelMetadata) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}
func (this *ModelMetadataDynamicRepository) Exists(ctx corectx.Context, keys []domain.ModelMetadata) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	dynamicKeys := array.Map(keys, func(key domain.ModelMetadata) dmodel.DynamicFields { return key.GetFieldData() })
	return baserepo.Exists(ctx, this.dynamicRepo, dynamicKeys)
}
func (this *ModelMetadataDynamicRepository) Insert(ctx corectx.Context, src domain.ModelMetadata) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, src)
}
func (this *ModelMetadataDynamicRepository) GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.ModelMetadata], error) {
	return baserepo.GetOne[domain.ModelMetadata](ctx, this.dynamicRepo, param)
}
func (this *ModelMetadataDynamicRepository) Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.ModelMetadata]], error) {
	return baserepo.Search[domain.ModelMetadata](ctx, this.dynamicRepo, param)
}
func (this *ModelMetadataDynamicRepository) Update(ctx corectx.Context, src domain.ModelMetadata) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, src.GetFieldData())
}
