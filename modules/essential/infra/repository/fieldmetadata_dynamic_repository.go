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
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/fieldmetadata"
)

type FieldMetadataDynamicRepositoryParam struct {
	dig.In

	Client        orm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  orm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewFieldMetadataDynamicRepository(param FieldMetadataDynamicRepositoryParam) it.FieldMetadataRepository {
	dynamicRepo := param.NewBaseRepoFn(
		dyn.NewBaseRepoParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(domain.FieldMetadataSchemaName),
		},
	)
	return &FieldMetadataDynamicRepository{dynamicRepo: dynamicRepo}
}

type FieldMetadataDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *FieldMetadataDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *FieldMetadataDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}

func (this *FieldMetadataDynamicRepository) DeleteOne(ctx corectx.Context, keys domain.FieldMetadata) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}

func (this *FieldMetadataDynamicRepository) Exists(ctx corectx.Context, keys []domain.FieldMetadata) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	dynamicKeys := array.Map(keys, func(key domain.FieldMetadata) dmodel.DynamicFields {
		return key.GetFieldData()
	})
	return baserepo.Exists(ctx, this.dynamicRepo, dynamicKeys)
}

func (this *FieldMetadataDynamicRepository) Insert(ctx corectx.Context, src domain.FieldMetadata) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, src)
}

func (this *FieldMetadataDynamicRepository) GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.FieldMetadata], error) {
	return baserepo.GetOne[domain.FieldMetadata](ctx, this.dynamicRepo, param)
}

func (this *FieldMetadataDynamicRepository) Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.FieldMetadata]], error) {
	return baserepo.Search[domain.FieldMetadata](ctx, this.dynamicRepo, param)
}

func (this *FieldMetadataDynamicRepository) Update(ctx corectx.Context, src domain.FieldMetadata) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, src.GetFieldData())
}
