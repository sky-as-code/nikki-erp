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
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/vendor"
)

type VendorDynamicRepositoryParam struct {
	dig.In
	Client        orm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  orm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewVendorDynamicRepository(param VendorDynamicRepositoryParam) it.VendorRepository {
	repo := param.NewBaseRepoFn(dyn.NewBaseRepoParam{
		Client: param.Client, ConfigSvc: param.ConfigSvc, QueryBuilder: param.QueryBuilder, Logger: param.Logger,
		Schema: dmodel.MustGetSchema(models.VendorSchemaName),
	})
	return &VendorDynamicRepository{dynamicRepo: repo}
}

type VendorDynamicRepository struct{ dynamicRepo dyn.BaseDynamicRepository }

func (this *VendorDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository { return this.dynamicRepo }
func (this *VendorDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}
func (this *VendorDynamicRepository) DeleteOne(ctx corectx.Context, keys models.Vendor) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}
func (this *VendorDynamicRepository) Exists(ctx corectx.Context, keys []models.Vendor) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	return baserepo.Exists(ctx, this.dynamicRepo, array.Map(keys, func(key models.Vendor) dmodel.DynamicFields {
		return key.GetFieldData()
	}))
}
func (this *VendorDynamicRepository) Insert(ctx corectx.Context, input models.Vendor) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, input)
}
func (this *VendorDynamicRepository) GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[models.Vendor], error) {
	return baserepo.GetOne[models.Vendor](ctx, this.dynamicRepo, param)
}
func (this *VendorDynamicRepository) Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[models.Vendor]], error) {
	return baserepo.Search[models.Vendor](ctx, this.dynamicRepo, param)
}
func (this *VendorDynamicRepository) Update(ctx corectx.Context, input models.Vendor) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, input.GetFieldData())
}
