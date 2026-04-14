package repository

import (
	"math"

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
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/module"
)

const moduleMetadataSyncLockKey = math.MaxInt64

type ModuleDynamicRepositoryParam struct {
	dig.In

	Client        orm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  orm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewModuleDynamicRepository(param ModuleDynamicRepositoryParam) it.ModuleRepository {
	dynamicRepo := param.NewBaseRepoFn(
		dyn.NewBaseRepoParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(domain.ModuleMetadataSchemaName),
		},
	)
	return &ModuleDynamicRepository{dynamicRepo: dynamicRepo}
}

type ModuleDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *ModuleDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *ModuleDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}

func (this *ModuleDynamicRepository) AcquireLock(ctx corectx.Context) (bool, error) {
	var acquired bool
	rows, err := this.dynamicRepo.QueryFunc(ctx, "pg_try_advisory_lock", moduleMetadataSyncLockKey)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&acquired)
	}
	return acquired, err
}

func (this *ModuleDynamicRepository) ReleaseLock(ctx corectx.Context) error {
	err := this.dynamicRepo.ExecFunc(ctx, "pg_advisory_unlock", moduleMetadataSyncLockKey)
	return err
}

func (this *ModuleDynamicRepository) DeleteOne(
	ctx corectx.Context, keys domain.ModuleMetadata,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}

func (this *ModuleDynamicRepository) Exists(
	ctx corectx.Context, keys []domain.ModuleMetadata,
) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	dynamicKeys := array.Map(keys, func(key domain.ModuleMetadata) dmodel.DynamicFields {
		return key.GetFieldData()
	})
	return baserepo.Exists(ctx, this.dynamicRepo, dynamicKeys)
}

func (this *ModuleDynamicRepository) Insert(
	ctx corectx.Context, module domain.ModuleMetadata,
) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, module)
}

func (this *ModuleDynamicRepository) GetOne(
	ctx corectx.Context, param dyn.RepoGetOneParam,
) (*dyn.OpResult[domain.ModuleMetadata], error) {
	return baserepo.GetOne[domain.ModuleMetadata](ctx, this.dynamicRepo, param)
}

func (this *ModuleDynamicRepository) Search(
	ctx corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[domain.ModuleMetadata]], error) {
	return baserepo.Search[domain.ModuleMetadata](ctx, this.dynamicRepo, param)
}

func (this *ModuleDynamicRepository) Update(
	ctx corectx.Context, module domain.ModuleMetadata,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, module.GetFieldData())
}
