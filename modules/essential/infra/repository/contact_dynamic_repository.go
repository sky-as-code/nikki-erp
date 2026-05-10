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
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/contact"
)

type ContactDynamicRepositoryParam struct {
	dig.In

	Client        orm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  orm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewContactDynamicRepository(param ContactDynamicRepositoryParam) it.ContactRepository {
	dynamicRepo := param.NewBaseRepoFn(
		dyn.NewBaseRepoParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(models.ContactSchemaName),
		},
	)
	return &ContactDynamicRepository{dynamicRepo: dynamicRepo}
}

type ContactDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *ContactDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *ContactDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}

func (this *ContactDynamicRepository) DeleteOne(ctx corectx.Context, keys models.Contact) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}

func (this *ContactDynamicRepository) Exists(ctx corectx.Context, keys []models.Contact) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	dynamicKeys := array.Map(keys, func(key models.Contact) dmodel.DynamicFields {
		return key.GetFieldData()
	})
	return baserepo.Exists(ctx, this.dynamicRepo, dynamicKeys)
}

func (this *ContactDynamicRepository) Insert(ctx corectx.Context, contact models.Contact) (
	*dyn.OpResult[int], error,
) {
	return baserepo.Insert(ctx, this.dynamicRepo, contact)
}

func (this *ContactDynamicRepository) GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (
	*dyn.OpResult[models.Contact], error,
) {
	return baserepo.GetOne[models.Contact](ctx, this.dynamicRepo, param)
}

func (this *ContactDynamicRepository) Search(
	ctx corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[models.Contact]], error) {
	return baserepo.Search[models.Contact](ctx, this.dynamicRepo, param)
}

func (this *ContactDynamicRepository) Update(ctx corectx.Context, contact models.Contact) (
	*dyn.OpResult[dyn.MutateResultData], error,
) {
	return baserepo.Update(ctx, this.dynamicRepo, contact.GetFieldData())
}
