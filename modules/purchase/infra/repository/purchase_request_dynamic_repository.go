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
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain"
	it "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/purchaserequest"
)

type PurchaseRequestDynamicRepositoryParam struct {
	dig.In

	Client        orm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  orm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewPurchaseRequestDynamicRepository(param PurchaseRequestDynamicRepositoryParam) it.PurchaseRequestRepository {
	dynamicRepo := param.NewBaseRepoFn(
		dyn.NewBaseRepoParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(domain.PurchaseRequestSchemaName),
		},
	)
	return &PurchaseRequestDynamicRepository{dynamicRepo: dynamicRepo}
}

type PurchaseRequestDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *PurchaseRequestDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *PurchaseRequestDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}

func (this *PurchaseRequestDynamicRepository) DeleteOne(
	ctx corectx.Context, keys domain.PurchaseRequest,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}

func (this *PurchaseRequestDynamicRepository) Exists(
	ctx corectx.Context, keys []domain.PurchaseRequest,
) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	dynamicKeys := array.Map(keys, func(key domain.PurchaseRequest) dmodel.DynamicFields {
		return key.GetFieldData()
	})
	return baserepo.Exists(ctx, this.dynamicRepo, dynamicKeys)
}

func (this *PurchaseRequestDynamicRepository) Insert(
	ctx corectx.Context, input domain.PurchaseRequest,
) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, input)
}

func (this *PurchaseRequestDynamicRepository) GetOne(
	ctx corectx.Context, param dyn.RepoGetOneParam,
) (*dyn.OpResult[domain.PurchaseRequest], error) {
	return baserepo.GetOne[domain.PurchaseRequest](ctx, this.dynamicRepo, param)
}

func (this *PurchaseRequestDynamicRepository) Search(
	ctx corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[domain.PurchaseRequest]], error) {
	return baserepo.Search[domain.PurchaseRequest](ctx, this.dynamicRepo, param)
}

func (this *PurchaseRequestDynamicRepository) Update(
	ctx corectx.Context, input domain.PurchaseRequest,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, input.GetFieldData())
}
