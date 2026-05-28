package repository

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/array"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/baserepo"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attributevalue"
)

type AttributeValueDynamicRepositoryParam struct {
	dig.In

	Client        orm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  orm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewAttributeValueDynamicRepository(param AttributeValueDynamicRepositoryParam) it.AttributeValueRepository {
	dynamicRepo := param.NewBaseRepoFn(
		dyn.NewBaseRepoParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(domain.AttributeValueSchemaName),
		},
	)
	return &AttributeValueDynamicRepository{dynamicRepo: dynamicRepo}
}

type AttributeValueDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *AttributeValueDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *AttributeValueDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}

func (this *AttributeValueDynamicRepository) DeleteOne(ctx corectx.Context, keys domain.AttributeValue) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.DeleteOne(ctx, this.dynamicRepo, keys.GetFieldData())
}

func (this *AttributeValueDynamicRepository) Exists(ctx corectx.Context, keys []domain.AttributeValue) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	dynamicKeys := array.Map(keys, func(key domain.AttributeValue) dmodel.DynamicFields {
		return key.GetFieldData()
	})
	return baserepo.Exists(ctx, this.dynamicRepo, dynamicKeys)
}

func (this *AttributeValueDynamicRepository) Insert(
	ctx corectx.Context, attributeValue domain.AttributeValue,
) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, attributeValue)
}

func (this *AttributeValueDynamicRepository) GetOne(
	ctx corectx.Context, param dyn.RepoGetOneParam,
) (*dyn.OpResult[domain.AttributeValue], error) {
	return baserepo.GetOne[domain.AttributeValue](ctx, this.dynamicRepo, param)
}

func (this *AttributeValueDynamicRepository) Search(
	ctx corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[domain.AttributeValue]], error) {
	return baserepo.Search[domain.AttributeValue](ctx, this.dynamicRepo, param)
}

func (this *AttributeValueDynamicRepository) Update(
	ctx corectx.Context, attributeValue domain.AttributeValue,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return baserepo.Update(ctx, this.dynamicRepo, attributeValue.GetFieldData())
}

func (this *AttributeValueDynamicRepository) GetIdsByVariantId(
	ctx corectx.Context, variantId model.Id,
) ([]model.Id, error) {
	client := this.dynamicRepo.ExtractClient(ctx)

	q := `SELECT ` + domain.VarAttrValRelFieldAttrValueId + `
	      FROM ` + dmodel.MustGetSchema(domain.VarAttrValRelSchemaName).TableName() + `
	      WHERE ` + domain.VarAttrValRelFieldVariantId + ` = $1`

	rows, err := client.Query(ctx.InnerContext(), q, variantId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []model.Id
	for rows.Next() {
		var id model.Id
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
