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
	it "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/purchaseorder"
)

type PurchaseOrderDynamicRepositoryParam struct {
	dig.In
	Client        orm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  orm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewPurchaseOrderDynamicRepository(param PurchaseOrderDynamicRepositoryParam) it.PurchaseOrderRepository {
	repo := param.NewBaseRepoFn(dyn.NewBaseRepoParam{
		Client: param.Client, ConfigSvc: param.ConfigSvc, QueryBuilder: param.QueryBuilder, Logger: param.Logger,
		Schema: dmodel.MustGetSchema(domain.PurchaseOrderSchemaName),
	})
	return &PurchaseOrderDynamicRepository{dynamicRepo: repo}
}

type PurchaseOrderDynamicRepository struct{ dynamicRepo dyn.BaseDynamicRepository }

func (this *PurchaseOrderDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}
func (this *PurchaseOrderDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}
func (this *PurchaseOrderDynamicRepository) DeleteOne(ctx corectx.Context, keys domain.PurchaseOrder) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}
func (this *PurchaseOrderDynamicRepository) Exists(ctx corectx.Context, keys []domain.PurchaseOrder) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	return baserepo.Exists(ctx, this.dynamicRepo, array.Map(keys, func(k domain.PurchaseOrder) dmodel.DynamicFields { return k.GetFieldData() }))
}
func (this *PurchaseOrderDynamicRepository) Insert(ctx corectx.Context, input domain.PurchaseOrder) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, input)
}
func (this *PurchaseOrderDynamicRepository) GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.PurchaseOrder], error) {
	return baserepo.GetOne[domain.PurchaseOrder](ctx, this.dynamicRepo, param)
}
func (this *PurchaseOrderDynamicRepository) Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.PurchaseOrder]], error) {
	return baserepo.Search[domain.PurchaseOrder](ctx, this.dynamicRepo, param)
}
func (this *PurchaseOrderDynamicRepository) Update(ctx corectx.Context, input domain.PurchaseOrder) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, input.GetFieldData())
}
