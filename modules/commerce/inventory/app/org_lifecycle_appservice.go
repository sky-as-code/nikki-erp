package app

import (
	"go.bryk.io/pkg/errors"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	orgDefaults "github.com/sky-as-code/nikki-erp/modules/commerce/infra/inventory/defaults"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/services"
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/warehouse"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle/defaults"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The locations a warehouse provisions for itself, which the operation type defaults name.
const (
	stockLocationCode     = "Stock"
	aliasWarehouseRootLoc = "location_main"
	aliasWarehouseStock   = "location_main_stock"
)

// cleanUpOrder lists every org-scoped resource children first. Attribute links without an
// org_id go with the templates and variants they cascade from.
var cleanUpOrder = []string{
	models.StockMoveDependencySchemaName, models.StockMoveLineSchemaName, models.StockMoveSchemaName,
	models.StockQuantSchemaName, models.StockReservationSchemaName, models.StockScrapSchemaName,
	models.ProductVariantSchemaName, models.StockProductConfigSchemaName, models.ProductTemplateSchemaName,
	models.BrandSchemaName, models.InventoryIntegrationOutboxSchemaName, models.PutawayRuleSchemaName,
	models.StockTransferSchemaName, models.StockOperationTypeSchemaName, models.InventoryLocationSchemaName,
	models.ProductAttributeValueSchemaName, models.ProductAttributeSchemaName,
	models.ProductCategorySchemaName, models.StorageCategorySchemaName,
	models.WarehouseProductGuardSchemaName, models.WarehouseSupplyRelationSchemaName, models.WarehouseSchemaName,
}

type OrgLifecycleParam struct {
	dig.In

	StorageCategory itWarehouse.StorageCategoryDomainService
	Category        itProduct.ProductCategoryDomainService
	Warehouse       itWarehouse.WarehouseAppService
	Location        itWarehouse.InventoryLocationDomainService
	OperationType   itStock.StockOperationTypeDomainService
}

func NewOrgLifecycleAppService(param OrgLifecycleParam) orglifecycle.OrgLifecycleService {
	return &OrgLifecycleAppServiceImpl{param: param, cleanup: composable.NewOrgDataCleanup(cleanUpOrder...)}
}

// OrgLifecycleAppServiceImpl gives a new org the root category, a main warehouse with the
// locations its flows need, the virtual locations every move starts or ends in, and the
// operation types sales and purchasing use.
type OrgLifecycleAppServiceImpl struct {
	param   OrgLifecycleParam
	cleanup composable.OrgDataCleanup
}

func (this *OrgLifecycleAppServiceImpl) InitOrg(ctx corectx.Context, orgId model.Id) error {
	return composable.InTransaction(ctx, this.param.Location.Repository(), func(tranxCtx corectx.Context) error {
		aliases := defaults.NewAliases()
		steps := []func(corectx.Context, *defaults.Aliases, model.Id) error{
			this.createStorageCategories, this.createRootCategory, this.createWarehouse,
			this.createVirtualLocations, this.createOperationTypes,
		}
		for _, step := range steps {
			if err := step(tranxCtx, aliases, orgId); err != nil {
				return err
			}
		}
		return nil
	})
}

func (this *OrgLifecycleAppServiceImpl) CleanUpOrg(ctx corectx.Context, orgId model.Id) error {
	return this.cleanup.CleanUpOrg(ctx, orgId)
}

func (this *OrgLifecycleAppServiceImpl) createStorageCategories(
	ctx corectx.Context, aliases *defaults.Aliases, orgId model.Id,
) error {
	return createFromDefaults(ctx, this.param.StorageCategory, aliases, orgId)
}

func (this *OrgLifecycleAppServiceImpl) createRootCategory(
	ctx corectx.Context, aliases *defaults.Aliases, orgId model.Id,
) error {
	rows, err := defaults.Load(orgDefaults.Files, models.ProductCategorySchemaName)
	if err != nil {
		return err
	}
	return aliases.CreateAll(rows, orgId, func(row dmodel.DynamicFields) (model.Id, error) {
		result, err := this.param.Category.CreateRoot(ctx, row)
		return createdId(result, err, models.ProductCategorySchemaName)
	})
}

// createWarehouse goes through the warehouse's own orchestration, which provisions the root and
// flow locations; their ids are remembered for the operation types.
func (this *OrgLifecycleAppServiceImpl) createWarehouse(
	ctx corectx.Context, aliases *defaults.Aliases, orgId model.Id,
) error {
	rows, err := defaults.Load(orgDefaults.Files, models.WarehouseSchemaName)
	if err != nil {
		return err
	}
	return aliases.CreateAll(rows, orgId, func(row dmodel.DynamicFields) (model.Id, error) {
		result, err := this.param.Warehouse.CreateWarehouse(ctx, itWarehouse.CreateWarehouseCommand{Fields: row})
		if err != nil {
			return "", err
		}
		if result.ClientErrors.Count() > 0 {
			return "", result.ClientErrors.ToError()
		}
		warehouseId := composable.ParamId(result.Data.Warehouse)
		code, _ := row[models.WarehouseFieldCode].(string)
		return warehouseId, rememberWarehouseLocations(ctx, aliases, string(warehouseId), code)
	})
}

func rememberWarehouseLocations(ctx corectx.Context, aliases *defaults.Aliases, warehouseId string, code string) error {
	for alias, locationCode := range map[string]string{aliasWarehouseRootLoc: code, aliasWarehouseStock: stockLocationCode} {
		found, err := services.FindWarehouseLocationByCode(ctx, warehouseId, locationCode)
		if err != nil {
			return err
		}
		if found == nil || found.GetId() == nil {
			return errors.Errorf("warehouse %s has no %s location", code, locationCode)
		}
		aliases.Remember(alias, *found.GetId())
	}
	return nil
}

// createVirtualLocations writes the org-wide locations outside any warehouse. Like the ones a
// warehouse provisions, they are system-generated, a flag the location service strips from input.
func (this *OrgLifecycleAppServiceImpl) createVirtualLocations(
	ctx corectx.Context, aliases *defaults.Aliases, orgId model.Id,
) error {
	rows, err := defaults.Load(orgDefaults.Files, models.InventoryLocationSchemaName)
	if err != nil {
		return err
	}
	return aliases.CreateAll(rows, orgId, func(row dmodel.DynamicFields) (model.Id, error) {
		id, err := composable.CreateOne(ctx, this.param.Location, row)
		if err != nil {
			return "", err
		}
		return id, services.MarkLocationSystemGenerated(ctx, string(id))
	})
}

func (this *OrgLifecycleAppServiceImpl) createOperationTypes(
	ctx corectx.Context, aliases *defaults.Aliases, orgId model.Id,
) error {
	return createFromDefaults(ctx, this.param.OperationType, aliases, orgId)
}

func createFromDefaults(
	ctx corectx.Context, svc composable.CrudDomainService, aliases *defaults.Aliases, orgId model.Id,
) error {
	rows, err := defaults.Load(orgDefaults.Files, svc.Schema().Name())
	if err != nil {
		return err
	}
	return composable.CreateDefaults(ctx, svc, rows, aliases, orgId)
}

func createdId(result *itProduct.CreateProductCategoryResult, err error, schemaName string) (model.Id, error) {
	if err != nil {
		return "", err
	}
	if result.ClientErrors.Count() > 0 {
		return "", result.ClientErrors.ToError()
	}
	if !result.HasData {
		return "", errors.Errorf("create %s returned no record", schemaName)
	}
	return composable.ParamId(result.Data), nil
}
